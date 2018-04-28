// Package protocols is where the protocol-testers live.
//
// Tests are dynamically instantiated at run-time, via a class-factory
// pattern, and due to their plugin nature they are simple to implement
// as they require only implementing a single method.
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
	// Arguments return the arguments which this protocol-test, along
	// with a regular expression which will be used to validate a non-empty
	// argument.
	//
	Arguments() map[string]string

	// Example should return a string describing how your protocol-test
	// works and is invoked.
	//
	// Optional arguments will automatically be documented.
	Example() string

	//
	//
	// RunTest actually invokes the protocol-handler to run its
	// tests.
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

// ProtocolHandler is the factory-method which looks up and returns
// an object of the given type - if possible.
func ProtocolHandler(id string) (a ProtocolTest) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor()
	}
	return
}

// Handlers returns the names of all the registered protocol-handlers.
func Handlers() []string {
	var result []string

	// For each handler save the name
	handlers.RLock()
	for index, _ := range handlers.m {
		result = append(result, index)
	}
	handlers.RUnlock()

	// And return the result
	return result

}
