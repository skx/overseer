// The notifier package is the thing that posts the results of
// executed tests into a Redis-queue.
//
// From here external processes are expected to fetch them,
// process them, and ultimately inform a human.
//
// The `bridges/` directory in the source-repository contains
// a couple of sample notifiers.
//


package notifier

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/skx/overseer/test"
)

// Notifier holds a connection to the Redis queue used to store test-results.
type Notifier struct {
	// Redis is the is the redis-handle
	Redis *redis.Client
}

// New is the constructor for the notifier, which copies the supplied redis
// handle into the object - and tests it for validity.
func New(client *redis.Client) (*Notifier, error) {

	// Create the object
	tmp := new(Notifier)

	// Save the redis-handle away
	tmp.Redis = client

	// And run a ping, just to make sure the handle
	// is open/valid.
	_, err := tmp.Redis.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		return nil, err
	}

	return tmp, nil
}

// Notify is the method which is invoked to send a notification
// via an addition to the set/queue on the redis host.
func (p *Notifier) Notify(test test.Test, result error) error {

	//
	// If we don't have a server configured then return immediately.
	//
	if p.Redis == nil {
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
	// Convert the MAP to a JSON string we can notify.
	//
	j, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to encode test-result to JSON: %s", err.Error())
		return err
	}

	//
	// Publish the message to the queue.
	//
	_, err = p.Redis.RPush("overseer.results", j).Result()
	if err != nil {
		fmt.Printf("Result addition failed: %s\n", err)
		return err
	}

	return nil
}
