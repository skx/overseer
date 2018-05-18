//
// This is the Purppura bridge, which reads test-results from redis, and submits
// them to purppura, such that a human can be notified of test failures.
//
// The program should be built like so:
//
//     go build purppura-bridge.go
//
// Once built launch it like so:
//
//     $ ./purppura-bridge -url="http://purppura.example.com/events"
//
// Every two minutes it will send a heartbeat to the purppura-server so
// that you know it is working.
//
// Steve
// --
//

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/robfig/cron"
)

// Avoid threading issues with our last update-time
var mutex sync.RWMutex

// The last time we received an update
var update int64

// Should we be verbose?
var verbose *bool

// The redis handle
var r *redis.Client

// The URL of the purppura server
var pURL *string

// The redis connection details
var redisHost *string
var redisPass *string

// Given a JSON string decode it and post to the Purppura URL.
func process(msg []byte) {

	// Update our last received time
	mutex.Lock()
	update = time.Now().Unix()
	mutex.Unlock()

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
	hash := hex.EncodeToString(hasher.Sum(nil))

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
	jsonValue, err := json.Marshal(values)
	if err != nil {
		fmt.Printf("process: Failed to encode JSON:%s\n", err.Error())
		os.Exit(1)
	}

	//
	// If we're being verbose show what we're going to POST
	//
	if *verbose {
		fmt.Printf("%s\n", jsonValue)
	}

	//
	// Post to purppura
	//
	res, err := http.Post(*pURL,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("process: Failed to post to purppura:%s\n", err.Error())
		os.Exit(1)
	}

	//
	// OK now we've submitted the post.
	//
	// We should retrieve the status-code + body, if the status-code
	// is "odd" then we'll show them.
	//
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("process: Error reading response to post: %s\n", err.Error())
		return
	}
	status := res.StatusCode

	if status != 200 {
		fmt.Printf("process: Error - Status code was not 200: %d\n", status)
		fmt.Printf("process: Response - %s\n", body)
	}
}

// CheckUpdates triggers an alert if we've not received anything recently
func CheckUpdates() {

	// Get our last-received time
	mutex.Lock()
	then := update
	mutex.Unlock()

	// Get the current time
	now := time.Now().Unix()

	//
	// The alert we'll send to the purppura server
	//
	values := map[string]string{
		"detail":  fmt.Sprintf("The purppura-bridge last received an update %d seconds ago.", now-then),
		"subject": "No traffic seen recently",
		"id":      "purppura-bridge-traffic",
	}

	// Raise or clear?
	if now-then > (60 * 5) {
		values["raise"] = "now"
	} else {
		values["raise"] = "clear"
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, err := json.Marshal(values)
	if err != nil {
		fmt.Printf("Failed to export to JSON - %s\n", err.Error())
		os.Exit(1)
	}

	//
	// Post to purppura
	//
	res, err := http.Post(*pURL,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("CheckUpdates: Failed to post purppura-bridge to purppura:%s\n", err.Error())
		os.Exit(1)
	}

	//
	// OK now we've submitted the post.
	//
	// We should retrieve the status-code + body, if the status-code
	// is "odd" then we'll show them.
	//
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("CheckUpdates: Error reading response to post: %s\n", err.Error())
		return
	}
	status := res.StatusCode

	if status != 200 {
		fmt.Printf("CheckUpdates: Error - Status code was not 200: %d\n", status)
		fmt.Printf("CheckUpdates: Response - %s\n", body)
	}
}

// SendHeartbeat updates the purppura server with a hearbeat alert.
// This will ensure that you're alerted if this bridge fails, dies, or
// isn't running
func SendHeartbeat() {

	//
	// The alert we'll send to the purppura server
	//
	values := map[string]string{
		"detail":  "The purppura-bridge hasn't sent a heartbeat recently, which means that overseer test-results won't raise alerts.",
		"subject": "The purppura bridge isn't running!",
		"id":      "purppura-bridge-heartbeat",
		"raise":   "+5m",
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, _ := json.Marshal(values)

	//
	// Post to purppura
	//
	res, err := http.Post(*pURL,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("SendHeartbeat: Failed to post heartbeat to purppura:%s\n", err.Error())
		os.Exit(1)
	}

	//
	// OK now we've submitted the post.
	//
	// We should retrieve the status-code + body, if the status-code
	// is "odd" then we'll show them.
	//
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("SendHeartbeat: Error reading response to post: %s\n", err.Error())
		return
	}
	status := res.StatusCode

	if status != 200 {
		fmt.Printf("SendHeartbeat: Error - Status code was not 200: %d\n", status)
		fmt.Printf("SendHeartbeat: Response - %s\n", body)
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
	pURL = flag.String("purppura", "", "The purppura-server URL")
	verbose = flag.Bool("verbose", false, "Be verbose?")
	flag.Parse()

	//
	// Sanity-check
	//
	if *pURL == "" {
		fmt.Printf("Usage: purppura-bridge -purpurra=https://alert.steve.fi/events [-redis-host=127.0.0.1:6379] [-redis-pass=secret]\n")
		os.Exit(1)

	}

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

	c := cron.New()
	// Make sure we send a heartbeat so we're alerted if the bridge fails
	c.AddFunc("@every 30s", func() { SendHeartbeat() })
	// Make sure we raise an alert if we don't have MQ-traffic
	c.AddFunc("@every 30s", func() { CheckUpdates() })
	c.Start()

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
