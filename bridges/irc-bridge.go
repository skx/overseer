package main

import (
	"encoding/json"

	"crypto/tls"
	"fmt"
	"net/url"
	"sync"

	"github.com/thoj/go-ircevent"

	"os"
	"os/signal"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// ircconn holds the IRC server connection.
var irccon *irc.Connection

// Have we joined our channel?
var joined bool

// Record the channel name here, for sending the message
var channel string

// Avoid threading issues with our passed/failed counts
var mutex sync.RWMutex

// Count of how many tests have executed and passed
var passed int64

// Count of how many tests have executed and failed
var failed int64

// The MQ handle
var mq *client.Client

// Given a JSON string decode and post to purppura
func process(msg []byte) {
	data := map[string]string{}

	if err := json.Unmarshal(msg, &data); err != nil {
		panic(err)
	}

	testType := data["type"]
	testTarget := data["target"]
	result := data["error"]

	//
	// Bump our pass/fail counts.
	//
	if result == "" {
		mutex.Lock()
		passed += 1
		mutex.Unlock()
	} else {
		mutex.Lock()
		failed += 1
		mutex.Unlock()
	}

	//
	// If the test passed then we don't care.
	//
	if result == "" {
		return
	}

	//
	// Format the failure message.
	//
	txt := fmt.Sprintf("The %s test against %s failed: %s", testType, testTarget, result)

	//
	// And send it.
	//
	irccon.Privmsg(channel, txt)
}

func setupIRC(data string) {

	//
	// Parse the configuration URL
	//
	u, err := url.Parse(data)
	if err != nil {
		panic(err)
	}

	//
	// Get fields.
	//
	irccon = irc.IRC(u.User.Username(), u.User.Username())

	//
	// Do we have a password?  If so set it.
	//
	pass, passPresent := u.User.Password()
	if passPresent && pass != "" {
		irccon.Password = pass
	}

	irccon.Debug = false

	//
	// We assum "irc://...." by default, but if ircs:// was
	// specified we'll allow TLS.
	//
	irccon.UseTLS = false
	if u.Scheme == "ircs" {
		irccon.UseTLS = true
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	//
	// Add a callback to join the channel
	//
	irccon.AddCallback("001", func(e *irc.Event) {
		channel = "#" + u.Fragment
		irccon.Join(channel)

		// Now we've joined
		joined = true
	})

	//
	// Because our connection is persistent we can use
	// it to process private messages.
	//
	// In this case we'll just say "No".
	//
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			//
			// event.Message() contains the message
			// event.Nick Contains the sender
			// event.Arguments[0] Contains the channel
			//
			// Send a private-reply.
			//
			mutex.Lock()
			var p = passed
			var f = failed
			mutex.Unlock()

			irccon.Privmsg(event.Nick,
				fmt.Sprintf("Total tests executed %d, %d passed, %d failed", p+f, p, f))
		}(event)
	})

	//
	// Connect
	//
	err = irccon.Connect(u.Host)
	if err != nil {
		panic(err)
	}

	for joined == false {

	}

}

func main() {

	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	addr := "127.0.0.1:1883"

	fmt.Printf("Connecting to IRC ..\n")
	setupIRC("irc://moi:@localhost:6667/#test")
	fmt.Printf("Connected to IRC ..\n")

	//
	// Create an MQTT Client.
	//
	mq = client.New(&client.Options{})

	//
	// Connect to the MQTT Server.
	//
	err := mq.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  addr,
		ClientID: []byte("overseer-watcher"),
	})
	if err != nil {
		fmt.Printf("Error connecting: %s\n", err.Error())
		os.Exit(1)
	}

	//
	// Subscribe to the channel
	//
	err = mq.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte("overseer"),
				QoS:         mqtt.QoS0,

				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					process(message)
				},
			},
		},
	})

	// Wait for receiving a signal.
	<-sigc

	// Disconnect the Network Connection.
	if err := mq.Disconnect(); err != nil {
		panic(err)
	}
}
