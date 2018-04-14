//
// This is our DNS protocol-test.
//
// It is more complex than the others, because it requires a complex
// invocation:
//
//   ns.example must run dns for hostname.example.com resolving A as '1.2.3.4'
//
//
package protocols

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/miekg/dns"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type DNSTest struct {
	input   string
	options TestOptions
}

//
// Here we have a map of types, we only cover the few we care about.
//
var StringToType = map[string]uint16{
	"A":    dns.TypeA,
	"AAAA": dns.TypeAAAA,
	"MX":   dns.TypeMX,
	"NS":   dns.TypeNS,
	"TXT":  dns.TypeTXT,
}

var (
	localm *dns.Msg
	localc *dns.Client
)

//
// lookup will perform a DNS query, using the nameservers in /etc/resolv.conf,
// and return an array of maps of the response
//
func (s *DNSTest) lookup(server string, name string, ltype string) ([]string, error) {

	var results []string

	var err error
	localm = &dns.Msg{
		MsgHdr: dns.MsgHdr{
			RecursionDesired: true,
		},
		Question: make([]dns.Question, 1),
	}
	localc = &dns.Client{
		ReadTimeout: s.options.Timeout,
	}
	r, err := s.localQuery(server, dns.Fqdn(name), ltype)
	if err != nil || r == nil {
		return nil, err
	}
	if r.Rcode == dns.RcodeNameError {
		return nil, errors.New(fmt.Sprintf("No such domain %s\n", dns.Fqdn(name)))
	}

	for _, ent := range r.Answer {

		//
		// Lookup the value
		//
		switch ent.(type) {
		case *dns.A:
			a := ent.(*dns.A).A
			results = append(results, fmt.Sprintf("%s", a))
		case *dns.AAAA:
			aaaa := ent.(*dns.AAAA).AAAA
			results = append(results, fmt.Sprintf("%s", aaaa))
		case *dns.MX:
			mx_name := ent.(*dns.MX).Mx
			mx_prio := ent.(*dns.MX).Preference
			results = append(results, fmt.Sprintf("%d %s", mx_prio, mx_name))
		case *dns.NS:
			nameserver := ent.(*dns.NS).Ns
			results = append(results, nameserver)
		case *dns.TXT:
			txt := ent.(*dns.TXT).Txt
			results = append(results, fmt.Sprintf("%s", txt[0]))
		}
	}
	return results, nil
}

//
// Given a thing to lookup, and a type, do the necessary.
//
// e.g. "steve.fi" "txt"
//
func (s *DNSTest) localQuery(server string, qname string, lookupType string) (*dns.Msg, error) {
	qtype := StringToType[lookupType]
	if qtype == 0 {
		return nil, errors.New(fmt.Sprintf("Unsupported record to lookup '%s'", lookupType))
	}
	localm.SetQuestion(qname, qtype)

	//
	// Default to connecting to an IPv4-address
	//
	address := fmt.Sprintf("%s:%d", server, 53)

	//
	// If we find a ":" we know it is an IPv6 address though
	//
	if strings.Contains(server, ":") {
		address = fmt.Sprintf("[%s]:%d", server, 53)
	}

	//
	// Run the lookup
	//
	r, _, err := localc.Exchange(localm, address)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Rcode == dns.RcodeNameError || r.Rcode == dns.RcodeSuccess {
		return r, err
	}
	return nil, nil
}

//
// Make a DNS-test.
//
func (s *DNSTest) RunTest(target string) error {

	//
	// Parse the options and make sure we have enough.
	//
	options := ParseArguments(s.input)

	if options["lookup"] == "" {
		return errors.New("No value to lookup specified.")
	}
	if options["type"] == "" {
		return errors.New("No record-type to lookup.")
	}

	//
	// NOTE:
	// "result" must also be specified, but it is valid to set that
	// to be empty.
	//

	//
	// Run the lookup
	//
	res, err := s.lookup(target, options["lookup"], options["type"])
	if err != nil {
		return err
	}

	//
	// If the results differ that's an error
	//
	// Sort the results and comma-join for comparison
	//
	sort.Strings(res)
	found := strings.Join(res, ",")

	if found == options["result"] {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Expected DNS result to be '%s', but found '%s'", options["result"], found))
	}
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific
// options to be understood.
//
func (s *DNSTest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *DNSTest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("dns", func() ProtocolTest {
		return &DNSTest{}
	})
}
