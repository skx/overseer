package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/skx/overseer/test"
)

// The handle to our server
var r *redis.Client

// Connect connects to the redis-server specified, and returns an error
// if that fails.
func ConnectResults(addr string, password string) error {

	r = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // use default DB
	})

	return nil
}

// MQNotify is the method which is invoked to send a notification
// via the MQ connection setup in `ConnectMQ`.
func NotifyResult(test test.Test, result error) error {

	//
	// If we don't have a server configured then return immediately.
	//
	if r == nil {
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
	// Publish the message to the queue.
	//
	_, err = r.RPush("overseer.results", j).Result()
	if err != nil {
		fmt.Printf("Result addition failed: %s\n", err)
		return err
	}

	return nil
}
