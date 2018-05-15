package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/skx/overseer/test"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// The MQ handle
var mq *client.Client

// ConnectMQ connects to the MQ-server specified, and returns an error
// if that fails.
func ConnectMQ(addr string) error {

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
		ClientID: []byte("overseer-client"),
	})
	if err != nil {
		return (err)
	}

	//
	// Ensure that our connection is established before we return
	//
	connected := false

	//
	// While we're not connected - try to subscribe to a topic
	//
	// The topic-name is irrelevant, we just want to make sure
	// that our async code is completed before we return from this
	// method.
	//
	for connected == false {
		err = mq.Subscribe(&client.SubscribeOptions{
			SubReqs: []*client.SubReq{
				&client.SubReq{
					TopicFilter: []byte("overseer"),
					QoS:         mqtt.QoS0,
					Handler: func(topicName, message []byte) {
					},
				},
			},
		})
		if err == nil {

			// If we didn't receive an error then we're
			// connected to our MQ server.

			connected = true
		}

		// Round and round we go.
		time.Sleep(10 * time.Millisecond)
	}

	//
	// Unsubscribe, now we've connected to the server.
	//
	err = mq.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte("overseer"),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// MQNotify is the method which is invoked to send a notification
// via the MQ connection setup in `ConnectMQ`.
func MQNotify(test test.Test, result error) error {

	//
	// The topic we'll publish to
	//
	topic := "overseer"

	//
	// If we don't have a server configured then return immediately.
	//
	if mq == nil {
		return nil
	}

	//
	// The message we'll publish will be a JSON hash
	//
	msg := map[string]string{
		"input":  test.Input,
		"result": "passed",
		"target": test.Target,
		"time":   fmt.Sprintf("%d", time.Now().Unix()),
		"type":   test.Type,
	}

	//
	// Was the test result a failure?  If so update the object
	// to contain the failure-message, and record that it was
	// a failure rather than the default pass.
	//
	if result != nil {
		msg["result"] = "failed"
		msg["error"] = result.Error()
	}

	//
	// Convert the MAP to a JSON string we can send down the MQ link.
	//
	j, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to encode test-result to JSON: %s", err.Error())
		return err
	}

	//
	// Rather than firing our message at the MQ host and then
	// returning we must wait until it is processed.
	//
	// Here we record whether we've received the message, proving
	// it has made it to the queue.
	//
	received := false

	//
	// Subscribe to the topic we're going to publish to.
	//
	// This is a horrid solution which is designed to ensure
	// that we don't terminate before the message has been
	// processed by the MQ library - and actually sent on the
	// queue.
	//
	// So here we're subscribing to the topic, and when we see
	// the message we've published ourself we can update our
	// `received` flag.
	//
	err = mq.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(topic),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {

					//
					// If the message received is
					// the one we are about to send
					// then it arrived and we're good
					//
					if string(message) == string(j) {
						received = true
					}
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to our topic: %s", err.Error())
		return err
	}

	//
	// Publish the message to the queue.
	//
	err = mq.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    true,
		TopicName: []byte(topic),
		Message:   j,
	})
	if err != nil {
		fmt.Printf("Publish to MQ failed: %s\n", err)
		return err
	}

	//
	// Now we're going to busy-wait such that we don't return
	// from this function unless/until we've seen our message
	// has made it to the queue.
	//
	// Now we're going to busy-wait until the message
	// has been received - via the subscription above
	//
	for received == false {
		time.Sleep(50 * time.Millisecond)
	}

	//
	// Unsubscribe from the topic, now we know our message was sent
	//
	err = mq.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
