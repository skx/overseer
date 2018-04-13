//
// This is our redis protocol-test.
//
//
package protocols

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type REDISTest struct {
	input   string
	options TestOptions
}

//
// Make a Redis-test against the given target.
//
func (s *REDISTest) RunTest(target string) error {

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
	// Default to connecting to an IPv4-address
	//
	address := fmt.Sprintf("%s:%d", target, port)

	//
	// If we find a ":" we know it is an IPv6 address though
	//
	if strings.Contains(target, ":") {
		address = fmt.Sprintf("[%s]:%d", target, port)
	}

	//
	// Attempt to connect to the host with the optional password
	//
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // use default DB
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
func (s *REDISTest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *REDISTest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("redis", func() ProtocolTest {
		return &REDISTest{}
	})
}
