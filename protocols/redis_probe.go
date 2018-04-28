// Redis Tester
//
// The Redis tester connects to a remote host and ensures that this succeeds,
// if a password is specified it will be used in the connection.
//
// This test is invoked via input like so:
//
//    host.example.com must run redis [with port 6379] [with password 'password']
//
package protocols

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/skx/overseer/test"
)

// REDISTest is our object
type REDISTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *REDISTest) Arguments() map[string]string {
	known := map[string]string{
		"port":     "^[0-9]+$",
		"password": ".*",
	}
	return known
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a Redis-test against the given target.
//
func (s *REDISTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {

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
	// If the user specified a different port update to use it.
	//
	if tst.Arguments["port"] != "" {
		port, err = strconv.Atoi(tst.Arguments["port"])
		if err != nil {
			return err
		}
	}

	//
	// If the user specified a password use it.
	//
	password = tst.Arguments["password"]

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
// Register our protocol-tester.
//
func init() {
	Register("redis", func() ProtocolTest {
		return &REDISTest{}
	})
}
