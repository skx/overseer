// IRC
//
// The IRC notification object sends test-failures over IRC
//
// Assuming you wish to post notification faiures to the room
// "#sysadmin", on the server "irc.example.com", as user
// "Bot" you'd set your connection string to:
//
//    irc://Bot@irc.example.com/#sysadmin
//
// If you wish to use TLS use "ircs":
//
//    ircs://Bot:password@irc.example.com/#sysadmin
//
// Finally if you need a password to join the server add ":password"
// appropriately.
//
package notifiers

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/skx/overseer/test"
	"github.com/thoj/go-ircevent"
)

// Our structure.
type IRCNotifier struct {
	options Options
}

// Given a configuration URL and a message, then send the message to the
// server.
func (s *IRCNotifier) SendIRCMessage(config string, message string) error {

	//
	// Parse the configuration URL
	//
	u, err := url.Parse(config)
	if err != nil {
		return err
	}

	//
	// Get fields.
	//
	irccon := irc.IRC(u.User.Username(), u.User.Username())

	//
	// Do we have a password?  If so set it.
	//
	pass, pass_present := u.User.Password()
	if pass_present && pass != "" {
		irccon.Password = pass
	}

	irccon.Debug = false

	//
	// We assum "irc://...." by default, but if ircs:// was
	// specified we'll allow TLS.
	//
	irccon.UseTLS = false
	if u.Scheme == "ircs" {
		irccon.UseTLS = true
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	//
	// Callback for joining
	//
	irccon.AddCallback("001", func(e *irc.Event) {
		channel := "#" + u.Fragment
		irccon.Join(channel)
		irccon.Privmsg(channel, message)
		irccon.Quit()
	})

	//
	// Connect
	//
	err = irccon.Connect(u.Host)
	if err != nil {
		return err
	}

	//
	// We'll loop only enough time for this message to be sent
	//
	irccon.Loop()

	return nil
}

// Send a notification to the IRC server.
func (s *IRCNotifier) Notify(test test.Test, result error) error {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if s.options.Data == "" {
		return nil
	}

	//
	// If the test passed then we don't care
	//
	if result == nil {
		return nil
	}

	//
	// OK so we have a) a test that failed, b) a configured
	// IRC connection-string
	//

	//
	// Format the failure message/
	//
	msg := fmt.Sprintf("The %s test against %s failed: %s", test.Type, test.Target, result.Error())

	//
	// And send it.
	//
	return (s.SendIRCMessage(s.options.Data, msg))
}

// Save the options we're given away
func (s *IRCNotifier) SetOptions(opts Options) {
	s.options = opts
}

// Register our notifier
func init() {
	Register("irc", func() Notifier {
		return &IRCNotifier{}
	})
}
