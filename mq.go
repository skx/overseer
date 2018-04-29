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

	return nil
}

// MQNotify is the method which is invoked to send a notification
// via the existing MQ connection.
func MQNotify(test test.Test, result error) error {

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
	// Was the result a failure?  If so update to have
	// the details.
	//
	if result != nil {
		msg["result"] = "failed"
		msg["error"] = result.Error()
	}

	//
	// Convert the MAP to a JSON hash
	//
	j, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to encode JSON")
		return err
	}

	//
	// Publish the message.
	//
	err = mq.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    true,
		TopicName: []byte("overseer"),
		Message:   j,
	})
	if err != nil {
		fmt.Printf("Publish to MQ failed: %s\n", err)
		return err
	}

	return nil
}
