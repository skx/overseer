// DNS Tester
//
// The DNS tester allows you to confirm that the specified DNS server
// returns the results you expect.  It is invoked with input like this:
//
//    ns.example.com must run dns with lookup test.example.com with type A with result '1.2.3.4'
//
// This test ensures that the DNS lookup of an A record for `test.example.com`
// returns the single value 1.2.3.4
//
// Lookups are supported for A, AAAA, MX, NS, and TXT records.
//

package protocols

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/skx/overseer/test"
	//	"github.com/skx/overseer/protocols"
)

// DNSTest is our object.
type DNSTest struct {
}

// Here we have a map of DNS type-names.
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

// lookup will perform a DNS query, using the servername-specified.
// It returns an array of maps of the response.
func (s *DNSTest) lookup(server string, name string, ltype string, timeout time.Duration) ([]string, error) {

	var results []string

	var err error
	localm = &dns.Msg{
		MsgHdr: dns.MsgHdr{
			RecursionDesired: true,
		},
		Question: make([]dns.Question, 1),
	}
	localc = &dns.Client{
		ReadTimeout: timeout,
	}
	r, err := s.localQuery(server, dns.Fqdn(name), ltype)
	if err != nil || r == nil {
		return nil, err
	}
	if r.Rcode == dns.RcodeNameError {
		return nil, fmt.Errorf("No such domain %s\n", dns.Fqdn(name))
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
			mxName := ent.(*dns.MX).Mx
			mxPrio := ent.(*dns.MX).Preference
			results = append(results, fmt.Sprintf("%d %s", mxPrio, mxName))
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

// Given a name & type to lookup perform the request against the named
// DNS-server.
func (s *DNSTest) localQuery(server string, qname string, lookupType string) (*dns.Msg, error) {
	qtype := StringToType[lookupType]
	if qtype == 0 {
		return nil, fmt.Errorf("Unsupported record to lookup '%s'", lookupType)
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

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *DNSTest) Arguments() map[string]string {

	known := map[string]string{
		"type":   "A|AAAA|MX|NS|TXT",
		"lookup": ".*",
		"result": ".*",
	}
	return known
}

// Make a DNS-test.
func (s *DNSTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {

	if tst.Arguments["lookup"] == "" {
		return errors.New("No value to lookup specified.")
	}
	if tst.Arguments["type"] == "" {
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
	res, err := s.lookup(target, tst.Arguments["lookup"], tst.Arguments["type"], opts.Timeout)
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

	if found == tst.Arguments["result"] {
		return nil
	} else {
		return fmt.Errorf("Expected DNS result to be '%s', but found '%s'", tst.Arguments["result"], found)
	}
}

// Register our protocol-tester.
func init() {
	Register("dns", func() ProtocolTest {
		return &DNSTest{}
	})
}
