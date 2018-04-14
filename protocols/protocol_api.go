package protocols

import (
	"sync"
	"time"
)

// TestOptions are options which are passed to
// ever test-handler, via `SetOptions`.
//
// The options can change the way the test operates.
type TestOptions struct {
	// Should failing tests be retried?
	Retry bool

	// Timeout for the single test, in seconds
	Timeout time.Duration

	// Should the test consider itself to be running verbosely?
	Verbose bool

	// Should the test probe IPv4 addresses?
	IPv4 bool

	// Should the test probe IPv6 addresses?
	IPv6 bool
}

// ProtocolTest interface is the core of our code, it
// defines the implementation methods which must be
// implemented to add a new protocol-test.
type ProtocolTest interface {
	//
	// Run the test against the given target.
	//
	// Return a suitable error if the test fails, or
	// nil to indicate it passed.
	//
	RunTest(target string) error

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

	// Set the options for this test.
	SetOptions(opts TestOptions)
}

//
// This is a map of known-tests.
//
var handlers = struct {
	m map[string]TestCtor
	sync.RWMutex
}{m: make(map[string]TestCtor)}

//
// A constructor-function.
//
type TestCtor func() ProtocolTest

//
// Register a test-type with a constructor.
//
func Register(id string, newfunc TestCtor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

//
// Lookup the given type and create an instance of it,
// if we can.
//
func ProtocolHandler(id string) (a ProtocolTest) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor()
	}
	return
}
