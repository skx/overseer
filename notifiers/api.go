// Package notifiers contains a simple notification API which allows the
// results of tests to be submitted "somewhere".
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
// It is expected they contain an URL, target, credentials, or similar.
//
type Options struct {
	// The data passed to the notifier, as a string.
	Data string
}

// Notifier is the interface that must be fulfilled by our notifiers.
type Notifier interface {
	// Setup allows any notifier-specific setup actions
	// to take place.  For example a notifier that posts
	// messages to a slack-channel might connect to the
	// server here, using that connection in the Notify() method
	// when it is time to actually trigger a notification.
	Setup() error

	// Raise an alert, via some mechanism
	Notify(test test.Test, result error) error
}

// This is a map of known notifier types, and their corresponding constructors.
var handlers = struct {
	m map[string]Ctor
	sync.RWMutex
}{m: make(map[string]Ctor)}

// This is the signature of a constructor-function which may be registered
// as a notifier.
type Ctor func(data string) Notifier

// Register a notifier object with the specified constructor function.
func Register(id string, newfunc Ctor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

// Lookup the given notification-type and create an instance of it,
// if we can.
func NotifierType(id string, data string) (a Notifier) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor(data)
	}
	return
}
