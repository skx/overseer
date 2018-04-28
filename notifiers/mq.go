// MQ
//
// The MQ notification object sends test-results to a topic named `overseer`
// in a given MQ instance.
//
// Set your connection string to:
//
//    mq.example.com:1883
//
// This has been tested with the following queue - https://mosquitto.org/
//

package notifiers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/skx/overseer/test"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// Our structure.
type MQNotifier struct {
	// The connection-string we were passed in the constructor
	data string

	// The MQ handle
	mq *client.Client
}

// Connect to our MQ queue.
func (s *MQNotifier) Setup() error {

	//
	// Create an MQTT Client.
	//
	s.mq = client.New(&client.Options{})

	//
	// Connect to the MQTT Server.
	//
	err := s.mq.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  s.data,
		ClientID: []byte("overseer-client"),
	})
	if err != nil {
		return (err)
	}

	return nil
}

// Send a notification to our queue.
func (s *MQNotifier) Notify(test test.Test, result error) error {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if s.data == "" {
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
	// Publish our message.
	//
	err = s.mq.Publish(&client.PublishOptions{
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

// Register our notifier
func init() {
	Register("mq", func(data string) Notifier {
		return &MQNotifier{data: data}
	})
}
