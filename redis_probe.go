//
// This is our redis protocol-test.
//
//
package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-redis/redis"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type REDISTest struct {
	input string
}

//
// Make a Redis-test against the given target.
//
func (s *REDISTest) runTest(target string) error {

	//
	// Predeclare our error
	//
	var err error

	//
	// The default port to connect to.
	//
	port := 6379

	//
	// The default password to use.
	//
	password := ""

	//
	// If the user specified a different port update it.
	//
	report := regexp.MustCompile("on\\s+port\\s+([0-9]+)")
	out := report.FindStringSubmatch(s.input)
	if len(out) == 2 {
		port, err = strconv.Atoi(out[1])
		if err != nil {
			return err
		}
	}

	//
	// If the user specified a password use it.
	//
	repass := regexp.MustCompile("with\\s+password\\s+'([^']+)'")
	out = repass.FindStringSubmatch(s.input)
	if len(out) == 2 {
		password = out[1]
	}

	//
	// Attempt to connect
	//
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", target, port),
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	//
	// And run a ping
	//
	// If the connection is refused, or the auth-details don't match
	// then we'll see that here.
	//
	_, err = client.Ping().Result()
	if err != nil {
		return err
	}

	//
	// If we reached here all is OK
	//
	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific
// options to be understood.
//
func (s *REDISTest) setLine(input string) {
	s.input = input
}

//
// Register our protocol-tester.
//
func init() {
	Register("redis", func() ProtocolTest {
		return &REDISTest{}
	})
}
