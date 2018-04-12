package protocols

import (
	"sync"
	"time"
)

//
// ProtocolTest interface is the core of our code, it
// defines the implementation methods which must be
// implemented to add a new protocol-test.
//
type ProtocolTest interface {
	//
	// Run the test against the given target.
	//
	// If the test passed nil is returned, otherwise a suitable
	// error object.
	//
	RunTest(target string) error

	//
	// This function is invoked with the complete line
	// from the parser.  This is useful as some tests might
	// wish to allow extra options to be specified.
	//
	// For example a test might say:
	//
	//   http://example.com/ must run http with content 'steve'
	//
	// There is no general purpose way to specify options, so the
	// test itself can look for option-flags it recognizes.
	//
	SetLine(input string)

	//
	// Set the timeout period for this test
	//
	SetTimeout(timeout time.Duration)
}

//
// This is a map of known-tests.
//
var abilities = struct {
	m map[string]AbilityCtor
	sync.RWMutex
}{m: make(map[string]AbilityCtor)}

//
// A constructor-function.
//
type AbilityCtor func() ProtocolTest

//
// Register a test-type with a constructor.
//
func Register(id string, newfunc AbilityCtor) {
	abilities.Lock()
	abilities.m[id] = newfunc
	abilities.Unlock()
}

//
// Lookup the given type and create an instance of it,
// if we can.
//
func ProtocolHandler(id string) (a ProtocolTest) {
	abilities.RLock()
	ctor, ok := abilities.m[id]
	abilities.RUnlock()
	if ok {
		a = ctor()
	}
	return
}
