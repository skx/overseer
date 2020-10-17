//
// This is the telegram bridge, which reads test-results from redis, and submits
// notices of failures to telegram, such that a human can be notified.
//
// The program should be built like so:
//
//     go build .
//
// Once built launch it like so:
//
//     $ ./telegram-bridge -token=xxxx -recipient=YYY
//
// Here `xxxx` is the token for the telegram bot API, and YYY is the UID
// of the user to message.
//
// Steve
// --
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// The redis handle
var r *redis.Client

// The telegram bot token
var token *string

// The recipient of the message
var recipient *string

// Given a JSON message decode it and post to telegram if it describes a
// failure.
func process(msg []byte) error {

	data := map[string]string{}

	if err := json.Unmarshal(msg, &data); err != nil {
		return err
	}

	// If the test passed we don't care
	if data["error"] == "" {
		return nil
	}

	testType := data["type"]
	testTarget := data["target"]
	input := data["input"]

	// Make the target a link, if it looks like one.
	if strings.HasPrefix(testTarget, "http") {
		testTarget = fmt.Sprintf("<a href=\"%s\">%s</a>", testTarget, testTarget)
	}

	// The message we send to the user.
	text := fmt.Sprintf("The <code>%s</code> test failed against %s.\n\n%s\n\nThe test was:\n<code>%s</code>", testType, testTarget, data["error"], input)

	//
	// Create the bot
	//
	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		return fmt.Errorf("error creating telegram bot with token '%s': %s", *token, err.Error())
	}

	//
	// Convert the recipient string to a number.
	//
	n := 0
	n, err = strconv.Atoi(*recipient)
	if err != nil {
		return fmt.Errorf("error converting user to notify %s to integer: %s", *recipient, err.Error())
	}

	//
	// Create the message.
	//
	message := tgbotapi.NewMessage(int64(n), text)
	message.ParseMode = tgbotapi.ModeHTML

	//
	// Send the message.
	//
	_, err = bot.Send(message)
	if err != nil {
		return fmt.Errorf("error sending message to user %s", err.Error())
	}

	//
	// All done
	//
	return nil
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
	token = flag.String("token", "", "The telegram bot token")
	recipient = flag.String("recipient", "", "The telegram user to notify")
	flag.Parse()

	//
	// Sanity-check
	//
	if *recipient == "" || *token == "" {
		fmt.Printf("Please set the telegram recipient and token.\n")
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

	for {

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
			err := process([]byte(msg[1]))
			if err != nil {
				fmt.Printf("error notifying user: %s\n", err.Error())
				return
			}
		}
	}
}
