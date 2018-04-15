// The notifiers API allows the results of tests to be submitted
// "somewhere".
//
// The notification mechanism uses a class-factory to instantiate
// a single specific notifier, at run-time.
//
package notifiers

import (
	"sync"

	"github.com/skx/overseer/test"
)

// NotifierOptions contain options that are passed to the selected
// notifier.
//
// It is expected they contain an URL, target, credential, or similar.
type Options struct {
	// The data passed to the notifier, as a string.
	Data string
}

// Notifier is the interface that must be fulfilled by our notifiers.
type Notifier interface {
	// Raise an alert, via some mechanism
	Notify(test test.Test, result error) error

	// Set the options for this notifier.
	SetOptions(opts Options)
}

// This is a map of known notifier types, and their corresponding constructors.
var handlers = struct {
	m map[string]Ctor
	sync.RWMutex
}{m: make(map[string]Ctor)}

// This is the signature of a constructor-function which may be registered
// as a notifier.
type Ctor func() Notifier

// Register a notifier object with the specified constructor function.
func Register(id string, newfunc Ctor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

// Lookup the given notification-type and create an instance of it,
// if we can.
func NotifierType(id string) (a Notifier) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor()
	}
	return
}
