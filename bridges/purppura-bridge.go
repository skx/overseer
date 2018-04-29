package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"

	"net/http"

	"fmt"
	"os"
	"os/signal"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

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
	_, err := http.Post("http://localhost:8080/events",
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("Failed to post to purppura:%s\n", err.Error())
	}

}

func main() {

	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	addr := "127.0.0.1:1883"

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
			&client.SubReq{
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
