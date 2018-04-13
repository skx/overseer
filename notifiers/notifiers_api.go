package notifiers

import (
	"sync"

	"github.com/skx/overseer/parser"
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
	Notify(test parser.Test, result error) error

	// Set the options for this notifier.
	SetOptions(opts Options)
}

//
// This is a map of known-tests.
//
var handlers = struct {
	m map[string]Ctor
	sync.RWMutex
}{m: make(map[string]Ctor)}

//
// A constructor-function.
//
type Ctor func() Notifier

//
// Register a test-type with a constructor.
//
func Register(id string, newfunc Ctor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

//
// Lookup the given type and create an instance of it,
// if we can.
//
func NotifierType(id string) (a Notifier) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor()
	}
	return
}
