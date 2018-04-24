// Mosquitto
//
// The mosquitto notification object sends test-results to a
// mosquitto topic named `overseer`.
//
// Set your connection string to:
//
//    mq.example.com:1883
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
	// Build it up, via a map.
	//
	var msg map[string]string
	msg = make(map[string]string)

	msg["target"] = test.Target
	msg["type"] = test.Type
	msg["input"] = test.Input
	msg["time"] = fmt.Sprintf("%d", time.Now().Unix())

	//
	// The rest result.
	//
	if result != nil {
		msg["result"] = "failed"
		msg["error"] = result.Error()
	} else {
		msg["result"] = "passed"
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
