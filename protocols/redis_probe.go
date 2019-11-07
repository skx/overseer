// Redis Tester
//
// The Redis tester connects to a remote host and ensures that this succeeds,
// if a password is specified it will be used in the connection.
//
// This test is invoked via input like so:
//
//    host.example.com must run redis [with port 6379] [with password 'password']
//
// If you wish you can test the size of a set/list, this is not quite as
// good as it might be as our argument-parsing is a little too strict.
//
// To make sure the list `steve` has no more than `1000` entries we
// would write:
//
//   localhost must run redis with list 'steve' with max_size '1000'
//
// Or the set `users`:
//
//   localhost must run redis with set 'members' with max_size '1000'
//
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
		"list":     ".*",
		"set":      ".*",
		"max_size": "^[0-9]+$",
		"password": ".*",
		"port":     "^[0-9]+$",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *REDISTest) Example() string {
	str := `
Redis Tester
------------
 The Redis tester connects to a remote host and ensures that this succeeds,
 if a password is specified it will be used in the connection.

 This test is invoked via input like so:

    host.example.com must run redis [with password 'secret'] [with port '6379']

 If you wish you can test the size of a set/list, this is not quite as
 good as it might be as our argument-parsing is a little too strict.

 To make sure the list 'steve' has no more than 1000 entries we
 would write:

   localhost must run redis with list 'steve' with max_size '1000'

 Or the set 'users':

   localhost must run redis with set 'members' with max_size '1000'
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a Redis-test against the given target.
//
func (s *REDISTest) RunTest(tst test.Test, target string, opts test.Options) error {

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
	// Maximum list-size, if specified
	//
	maxSize := 0
	if tst.Arguments["max_size"] != "" {
		maxSize, err = strconv.Atoi(tst.Arguments["max_size"])
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
	// Now test the connection by running a ping
	//
	// If the connection is refused, or the auth-details don't match
	// then we'll see that here.
	//
	_, err = client.Ping().Result()
	if err != nil {
		return err
	}

	//
	// If we have a `list` and a `max_size` then get the size of the
	// specified set.
	//
	if tst.Arguments["list"] != "" && maxSize > 0 {

		//
		// Get the length of the list.
		//
		res := client.LLen(tst.Arguments["list"])
		if res.Err() != nil {
			return res.Err()
		}

		len := int(res.Val())

		//
		// Raise an alert if the size is exceeded.
		//
		if len >= maxSize {
			return (fmt.Errorf("list %s has %d entries, more than the max size of %d", tst.Arguments["list"], len, maxSize))
		}
	}

	//
	// If we have a `set` and a `max_size` then get the size of the
	// specified set.
	//
	if tst.Arguments["set"] != "" && maxSize > 0 {

		//
		// Get the count of set-members.
		//
		res := client.SCard(tst.Arguments["list"])
		if res.Err() != nil {
			return res.Err()
		}

		len := int(res.Val())

		//
		// Raise an alert if the size is exceeded.
		//
		if len >= maxSize {
			return (fmt.Errorf("set %s has %d members, more than the max size of %d", tst.Arguments["list"], len, maxSize))
		}
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
