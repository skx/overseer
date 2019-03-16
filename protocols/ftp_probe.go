// FTP Tester
//
// The FTP tester allows you to make a connection to an FTP-server,
// and optionally retrieve a file.
//
// A basic test can be invoked via input like so:
//
//    host.example.com must run ftp [with port 21]
//
// A more complex test would involve actually retrieving a file.  To make
// the test-definition natural you do this by specifying an URI:
//
//   ftp://ftp.cpan.org/pub/gnu/=README must run ftp
//
// Downloading a file requires a login, so by default we'll try an anonymous
// one.  If you need to specify real credentials you can do so by adding
// the appropriate username & password:
//
//   ftp://ftp.example.com/path/to/README must run ftp with username 'user@host.com' with password 'secret'
//
// Of course the URI could also be used to specify the login details:
//
//   ftp://user@example.com:secret@ftp.cpan.org/pub/gnu/=README must run ftp
//
// To ensure that the remote-file contains content you expect you can
// also verify a specific string is included within the response, via the
// "content" parameter:
//
//    ftp://ftp.example.com/path/to/README.md must run ftp with content '2018'

package protocols

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/jlaffaye/ftp"
	"github.com/skx/overseer/test"
)

// FTPTest is our object.
type FTPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *FTPTest) Arguments() map[string]string {
	known := map[string]string{
		"content":  ".*",
		"password": ".*",
		"port":     "^[0-9]+$",
		"username": ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *FTPTest) Example() string {
	str := `
FTP Tester
----------
 The FTP tester allows you to make a connection to an FTP-server,
 and optionally retrieve a file.

 A basic test can be invoked via input like so:

    host.example.com must run ftp [with port 21]

 A more complex test would involve actually retrieving a file.  To make
 the test-definition natural you do this by specifying an URI:

   ftp://ftp.cpan.org/pub/gnu/=README must run ftp

 Downloading a file requires a login, so by default we'll try an anonymous
 one.  If you need to specify real credentials you can do so by adding
 the appropriate username & password:

   ftp://ftp.example.com/path/to/README must run ftp with username 'user@host.com' with password 'secret'

 Of course the URI could also be used to specify the login details:

   ftp://user@example.com:secret@ftp.cpan.org/pub/gnu/=README must run ftp

 To ensure that the remote-file contains content you expect you can
 also verify a specific string is included within the response, via the
 "content" parameter:

    ftp://ftp.example.com/path/to/README.md must run ftp with content '2018'
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 21, and
// look for a response which appears to be an FTP-server.
func (s *FTPTest) RunTest(tst test.Test, target string, opts test.Options) error {
	//
	// Holder for any error we might encounter.
	//
	var err error

	//
	// The default port to connect to.
	//
	port := 21

	//
	// Our default credentials
	//
	username := "anonymous"
	password := "overseer@example.com"

	//
	// The target-file we're going to retrieve, if any
	//
	file := "/"

	//
	// If we've been given an URI then we should update the
	// port if it is non-standard, and possibly retrieve an
	// actual file too.
	//
	if strings.Contains(tst.Target, "://") {

		// Parse the URI.
		u, err := url.Parse(tst.Target)
		if err != nil {
			return err
		}

		// Record the path to fetch.
		file = u.Path

		// Update the default port, if a port-number was given.
		if u.Port() != "" {
			port, err = strconv.Atoi(u.Port())
			if err != nil {
				return err
			}
		}

		// The URI might contain username/password
		if u.User.Username() != "" {
			username = u.User.Username()
			p, _ := u.User.Password()
			if p != "" {
				password = p
			}
		}
	}

	fmt.Printf("Username: %s -> %s\n", username, password)
	//
	// If the user specified a different port update to use it.
	//
	// Do this after the URI-parsing.
	//
	if tst.Arguments["port"] != "" {
		port, err = strconv.Atoi(tst.Arguments["port"])
		if err != nil {
			return err
		}
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
	// Make the connection.
	//
	var conn *ftp.ServerConn
	conn, err = ftp.DialTimeout(address, opts.Timeout)
	if err != nil {
		return err
	}
	defer conn.Quit()

	//
	// If the user specified different/real credentials, use them instead.
	//
	if tst.Arguments["username"] != "" {
		username = tst.Arguments["username"]
	}
	if tst.Arguments["password"] != "" {
		password = tst.Arguments["password"]
	}

	//
	// If we have been given a path/file to fetch, via an URI
	// input, then fetch it.
	//
	// Before attempting the fetch login.
	//
	if file != "/" {

		//
		// Login
		//
		err = conn.Login(username, password)
		if err != nil {
			return err
		}

		//
		// Retrieve the file.
		//
		resp, err := conn.Retr(file)
		if err != nil {
			return err
		}
		defer resp.Close()

		//
		// Actually fetch the contents of the file.
		//
		buf, err := ioutil.ReadAll(resp)
		if err != nil {
			return err
		}

		//
		// If we're doing a content-match then do that here
		//
		if tst.Arguments["content"] != "" {
			if !strings.Contains(string(buf), tst.Arguments["content"]) {
				return fmt.Errorf("Body didn't contain '%s'", tst.Arguments["content"])
			}
		}

	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("ftp", func() ProtocolTest {
		return &FTPTest{}
	})
}
