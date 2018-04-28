// Purppura
//
// The purppura notification class is responsible for sending
// the results of any executed tests to an instance of the
// purppura notification system.
//
// The single argument is assumed to be the HTTP URL of the end-point
// to which submissions should be sent via HTTP POSTs.

package notifiers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skx/overseer/test"
)

// Purppura is our object
type Purppura struct {
	data string
}

// Setup is a NOP in this notification-class
func (s *Purppura) Setup() error {
	return nil
}

// Notify is the API-method which is invoked to send a notification
// somewhere.
//
// In our case we send a notification to the IRC server.
func (s *Purppura) Notify(test test.Test, result error) error {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if s.data == "" {
		return nil
	}

	testType := test.Type
	testTarget := test.Target
	input := test.Input

	//
	// We need a stable ID for each test - get one by hashing the
	// complete input-line and the target we executed against.
	//
	hasher := sha1.New()
	hasher.Write([]byte(testTarget))
	hasher.Write([]byte(input))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	//
	// All alerts will have an ID + Subject field.
	//
	values := map[string]string{
		"id":      hash,
		"subject": input,
	}

	//
	// If the test failed we'll set the detail + trigger a raise
	//
	if result != nil {
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> failed:</p><p><pre>%s</pre></p>",
				testType, testTarget, result.Error())
		values["raise"] = "now"
	} else {
		//
		// Otherwise the test passed and so all is OK
		//
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> passed.</p>",
				testType, testTarget)
		values["raise"] = "clear"
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, _ := json.Marshal(values)

	//
	// Post to purppura
	//
	_, err := http.Post(s.data,
		"application/json",
		bytes.NewBuffer(jsonValue))

	return err
}

// init is invoked to register our notifier at run-time.
func init() {
	Register("purppura", func(data string) Notifier {
		return &Purppura{data: data}
	})
}
