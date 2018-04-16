// The protocols package is where the protocol-testers live.
//
// Tests are dynamically instantiated at run-time, via a class-factory
// pattern, and due to their plugin nature they are simple to implement
// as they require only implementing a single method.
//
// There should now follow documentation on each available protocol-test.
//
package protocols

import (
	"sync"

	"github.com/skx/overseer/test"
)

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
	RunTest(tst test.Test, target string, opts test.TestOptions) error
}

// This is a map of known-tests.
var handlers = struct {
	m map[string]TestCtor
	sync.RWMutex
}{m: make(map[string]TestCtor)}

// TestCtor is the signature of a constructor-function.
type TestCtor func() ProtocolTest

// Register a test-type with a constructor.
func Register(id string, newfunc TestCtor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

// Lookup the given type of test and create an instance of it,
// if we can.
func ProtocolHandler(id string) (a ProtocolTest) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor()
	}
	return
}
