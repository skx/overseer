//
// This is the IRC bridge, which should be built via:
//
//     go build .
//
// Once built launch it like so:
//
//     $ ./irc-bridge -irc='irc://username:password@localhost:6667/#test'
//
// This will connect to the IRC server on localhost, with username "username"
// password "password", and post messages to "#test".
//
// Steve
// --
//

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"

	irc "github.com/thoj/go-ircevent"

	"github.com/go-redis/redis"
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

// The redis handle
var r *redis.Client

// The redis connection details
var redisHost *string
var redisPass *string

//
// Given a JSON string decode it and post to IRC if it describes
// a test-failure.
//
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
		passed++
		mutex.Unlock()
	} else {
		mutex.Lock()
		failed++
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

// setupIRC connects to the IRC server described by the specified URL.
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

	//
	// We don't need debugging information.
	//
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
	// In this case we will output statistics when private-messaged.
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
				fmt.Sprintf("Total tests executed %d\n", p+f))
			irccon.Privmsg(event.Nick,
				fmt.Sprintf("Failed tests %d\n", f))
			irccon.Privmsg(event.Nick,
				fmt.Sprintf("Succeeded tests %d\n", p))

		}(event)
	})

	//
	// Connect
	//
	err = irccon.Connect(u.Host)
	if err != nil {
		panic(err)
	}

	//
	// Wait until we've connected before returning.
	//
	for joined == false {
	}
}

//
// Entry Point
//
func main() {

	//
	// Parse our flags
	//
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "Specify the address of the redis queue.")
	redisPass := flag.String("redis-pass", "", "Specify the password of the redis queue.")
	irc := flag.String("irc", "", "A URL describing your IRC server")
	flag.Parse()

	//
	// Sanity-check.
	//
	if *irc == "" {
		fmt.Printf("Usage: irc-bridge -irc=irc://user:pass@irc.example.com:6667/#channel [-redis-host=127.0.0.1:6379] [-redis-pass=foo]\n")
		os.Exit(1)
	}

	//
	// Connect to IRC
	//
	fmt.Printf("Connecting to IRC server via %s ..\n", *irc)
	setupIRC(*irc)
	fmt.Printf("Connected.  Press Ctrl-c to terminate.\n")

	//
	// Create the redis client
	//
	r = redis.NewClient(&redis.Options{
		Addr:     *redisHost,
		Password: *redisPass,
		DB:       0, // use default DB
	})

	//
	// And run a ping, just to make sure it worked.
	//
	_, err := r.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		os.Exit(1)
	}

	for true {

		//
		// Get test-results
		//
		msg, _ := r.BLPop(0, "overseer.results").Result()

		//
		// If they were non-empty, process them.
		//
		//   msg[0] will be "overseer.results"
		//
		//   msg[1] will be the value removed from the list.
		//
		if len(msg) >= 1 {
			process([]byte(msg[1]))
		}
	}
}
