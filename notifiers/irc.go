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
	// data is the URI describing the IRC server to connect to
	data string

	// ircconn holds the IRC server connection.
	irccon *irc.Connection

	// Have we joined our channel?
	joined bool

	// Record the channel name here, for sending the message
	channel string
}

// Setup connects to the IRC server which was mentioned in the
// data passed to the constructor.
func (s *IRCNotifier) Setup() error {

	//
	// Parse the configuration URL
	//
	u, err := url.Parse(s.data)
	if err != nil {
		return err
	}

	//
	// Get fields.
	//
	s.irccon = irc.IRC(u.User.Username(), u.User.Username())

	//
	// Do we have a password?  If so set it.
	//
	pass, pass_present := u.User.Password()
	if pass_present && pass != "" {
		s.irccon.Password = pass
	}

	s.irccon.Debug = false

	//
	// We assum "irc://...." by default, but if ircs:// was
	// specified we'll allow TLS.
	//
	s.irccon.UseTLS = false
	if u.Scheme == "ircs" {
		s.irccon.UseTLS = true
		s.irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	//
	// Add a callback to join the channel
	//
	s.irccon.AddCallback("001", func(e *irc.Event) {
		s.channel = "#" + u.Fragment
		s.irccon.Join(s.channel)

		// Now we've joined
		s.joined = true
	})

	//
	// Because our connection is persistent we can use
	// it to process private messages.
	//
	// In this case we'll just say "No".
	//
	s.irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			//
			// event.Message() contains the message
			// event.Nick Contains the sender
			// event.Arguments[0] Contains the channel
			//
			// Send a private-reply.
			//
			s.irccon.Privmsg(event.Nick,
				"I don't accept private messages :)")
		}(event)
	})

	//
	// Connect
	//
	err = s.irccon.Connect(u.Host)
	if err != nil {
		return err
	}

	return nil
}

// Send a notification to the IRC server.
func (s *IRCNotifier) Notify(test test.Test, result error) error {

	//
	// If we don't have a server configured then return without sending
	// anything - there's no alternative since we don't know
	// which server/channel to use.
	//
	if s.data == "" {
		return nil
	}

	//
	// If the test passed then we don't care.
	//
	if result == nil {
		return nil
	}

	//
	// Format the failure message.
	//
	msg := fmt.Sprintf("The %s test against %s failed: %s", test.Type, test.Target, result.Error())

	//
	// And send it.
	//
	if s.joined {
		s.irccon.Privmsg(s.channel, msg)
	} else {
		fmt.Printf("Sending message before IRC server joined!")
	}

	return nil
}

// Register our notifier
func init() {
	Register("irc", func(data string) Notifier {

		return &IRCNotifier{data: data, joined: false}
	})
}
