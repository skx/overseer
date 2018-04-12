package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type SSHTest struct {
	input string
}

//
// Run the test against the specified target.
//
func (s *SSHTest) runTest(target string) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 22

	//
	// If the user specified a different port update it.
	//
	re := regexp.MustCompile("on\\s+port\\s+([0-9]+)")
	out := re.FindStringSubmatch(s.input)
	if len(out) == 2 {
		port, err = strconv.Atoi(out[1])
		if err != nil {
			return err
		}
	}

	//
	// Set an explicit timeout
	//
	d := net.Dialer{Timeout: time.Second * 10}

	//
	// Make the TCP connection.
	//
	conn, err := d.Dial("tcp", fmt.Sprintf("%s:%d", target, port))
	if err != nil {
		return err
	}

	//
	// Read the banner.
	//
	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	conn.Close()

	if !strings.Contains(banner, "OpenSSH") {
		return errors.New("Banner doesn't look like OpenSSH")
	}

	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *SSHTest) setLine(input string) {
	s.input = input
}

//
// Register our protocol-tester.
//
func init() {
	Register("ssh", func() ProtocolTest {
		return &SSHTest{}
	})
}
