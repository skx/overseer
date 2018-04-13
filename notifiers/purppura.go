package notifiers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skx/overseer/parser"
)

// Our structure.
type Purppura struct {
	options Options
}

// Send a notification
func (s *Purppura) Notify(test parser.Test, result error) error {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if s.options.Data == "" {
		return nil
	}

	test_type := test.Type
	test_target := test.Target
	input := test.Input

	//
	// We need a stable ID for each test - get one by hashing the
	// complete input-line and the target we executed against.
	//
	hasher := sha1.New()
	hasher.Write([]byte(test_target))
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
				test_type, test_target, result.Error())
		values["raise"] = "now"
	} else {
		//
		// Otherwise the test passed and so all is OK
		//
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> passed.</p>",
				test_type, test_target)
		values["raise"] = "clear"
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, _ := json.Marshal(values)

	//
	// Post to purppura
	//
	_, err := http.Post(s.options.Data,
		"application/json",
		bytes.NewBuffer(jsonValue))

	return err
}

// Save the options we're given away
func (s *Purppura) SetOptions(opts Options) {
	s.options = opts
}

// Register our notifier
func init() {
	Register("purppura", func() Notifier {
		return &Purppura{}
	})
}
