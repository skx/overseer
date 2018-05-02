//
// This is the Purppura bridge
//
// The program should be built via:
//
//     go build purppura-bridge.go
//
// Once built launch it like so:
//
//     $ ./purppura-bridge -mq="mq.example.com:1883" -url="http://purppura.example.com/events"
//
// This will connect to the MQ server specified and post to Purppura with
// the given URL.
//
// Steve
// --
//

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// The MQ handle
var mq *client.Client

// The URL of the purppura server
var pURL *string

// Given a JSON string decode it and post to the Purppura URL.
func process(msg []byte) {
	data := map[string]string{}

	if err := json.Unmarshal(msg, &data); err != nil {
		panic(err)
	}

	testType := data["type"]
	testTarget := data["target"]
	input := data["input"]

	//
	// We need a stable ID for each test - get one by hashing the
	// complete input-line and the target we executed against.
	//
	hasher := sha1.New()
	hasher.Write([]byte(testTarget))
	hasher.Write([]byte(input))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	//
	// Populate the default fields.
	//
	values := map[string]string{
		"detail":  fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> passed.</p>", testType, testTarget),
		"id":      hash,
		"raise":   "clear",
		"subject": input,
	}

	//
	// If the test failed we'll update the detail and trigger a raise
	//
	if data["error"] != "" {
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> failed:</p><p><pre>%s</pre></p>",
				testType, testTarget, data["error"])
		values["raise"] = "now"
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, _ := json.Marshal(values)

	//
	// Post to purppura
	//
	_, err := http.Post(*pURL,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("Failed to post to purppura:%s\n", err.Error())
	}

}

//
// Entry Point
//
func main() {

	//
	// Parse our flags
	//
	mqAddress := flag.String("mq", "127.0.0.1:1883", "The address & port of your MQ-server")
	pURL = flag.String("purppura", "", "The purppura-server URL")
	flag.Parse()

	//
	// Sanity-check
	//
	if *pURL == "" {
		fmt.Printf("Usage: purppura-bridge -mq=1.2.3.4:1883 -purpurra=https://alert.steve.fi/events\n")
		os.Exit(1)

	}

	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	//
	// Create an MQTT Client.
	//
	mq = client.New(&client.Options{})

	//
	// Connect to the MQTT Server.
	//
	err := mq.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  *mqAddress,
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
