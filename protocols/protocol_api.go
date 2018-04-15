package protocols

import (
	"sync"
	"time"

	"github.com/skx/overseer/test"
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
	// Run the specified test against the given target.
	//
	// Return a suitable error if the test fails, or nil to indicate
	// it passed.
	//
	RunTest(tst test.Test, target string, opts TestOptions) error
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
