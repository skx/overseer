package protocols

import (
	"bytes"
	"errors"
	"net"
	"os/exec"
	"syscall"

	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type PINGTest struct {
}

//
// Run a command, and return stdout/stderr/exit-code
//
func (s *PINGTest) RunCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = 1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return
}

// Ping4 runs a ping test against an IPv4 address, returning true
// if the ping succeeded.
func (s *PINGTest) Ping4(target string) bool {

	_, _, ret := s.RunCommand("ping4", "-c", "1", "-w", "4", "-W", "4", target)
	return (ret == 0)
}

// Ping6 runs a ping test against an IPv6 address, returning true
// if the ping succeeded.
func (s *PINGTest) Ping6(target string) bool {
	_, _, ret := s.RunCommand("ping6", "-c", "1", "-w", "4", "-W", "4", target)
	return (ret == 0)
}

// Return the arguments which this protocol-test understands.
func (s *PINGTest) Arguments() []string {
	known := []string{}
	return known
}

//
// Run the test against the specified target.
//
func (s *PINGTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	ip := net.ParseIP(target)

	//
	// If the address is an IPv4 address.
	//
	if ip.To4() != nil {
		if s.Ping4(target) {
			return nil
		}
		return errors.New("Failed to ping")
	}

	//
	// If the address is an IPv6 address.
	//
	if ip.To16() != nil && ip.To4() == nil {
		if s.Ping6(target) {
			return nil
		}
		return errors.New("Failed to ping")
	}

	//
	// Unknown family, or otherwise bogus name.
	//
	return errors.New("Neither IPv4 nor IPv6 address!")
}

//
// Register our protocol-tester.
//
func init() {
	Register("ping", func() ProtocolTest {
		return &PINGTest{}
	})
}
