package main

import (
	"bytes"
	"testing"

	"github.com/awgh/bencrypt/bc"
	"github.com/awgh/ratnet-transports/dns"
)

func Test_Dotify_Undotify_1(t *testing.T) {

	for i := 0; i < 151; i++ {
		testcase, err := bc.GenerateRandomBytes(i)
		if err != nil {
			t.Error(err.Error())
		}

		dot, err := dns.Dotify(testcase)
		if err != nil {
			t.Error(err.Error())
		}
		undot, err := dns.Undotify(dot)
		if err != nil {
			t.Error(err.Error())
		}

		if !bytes.Equal(testcase, undot) {
			t.Error("Equality check failed: ", testcase, len(testcase), undot, len(undot))
		}
	}
}

func Test_Dotify_Undotify_2(t *testing.T) {

	badbytes := []byte{68, 51, 34, 17, 81, 0, 32, 0, 3, 29, 94, 234, 3, 0, 0, 0, 4, 0, 0, 0, 42, 0, 0, 0, 79, 90, 112, 80, 53, 122, 57, 105, 85, 74, 114, 56, 83, 80, 75, 83, 98, 68, 76, 81, 114, 48, 110, 76, 98, 102, 75, 115, 72, 71, 106, 48, 118, 72, 110, 68, 72, 113, 103, 69, 61, 0}

	// base32 encode, then dotify / "DNS chop"
	dot, err := dns.Dotify(badbytes)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(dot)
	undot, err := dns.Undotify(dot)
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(badbytes, undot) {
		t.Error("Equality check failed: ", badbytes, len(badbytes), undot, len(undot))
	}
}

func Test_Add_RemoveDomainAGAIN(t *testing.T) {
	type testCase struct {
		domain    string
		subdomain string
		fqdn      string
	}
	testCases := []testCase{
		{
			"",
			"",
			"",
		},
		{
			"",
			"subdomain.",
			"subdomain.",
		},
		{
			"domain",
			"subdomain.",
			"subdomain.domain.",
		},
		{
			"domain.tld",
			"subdomain.",
			"subdomain.domain.tld.",
		},
		{
			"domain.tld",
			"subsubdomain.subdomain.",
			"subsubdomain.subdomain.domain.tld.",
		},
	}

	for _, tc := range testCases {
		fqdn := dns.AddDomain(tc.domain, tc.subdomain)
		if fqdn != tc.fqdn {
			t.Error("Equality check failed:", fqdn, tc.fqdn)
		}
		subdomain := dns.RemoveDomain(tc.domain, fqdn)
		if subdomain != tc.subdomain {
			t.Error("Equality check failed:", subdomain, tc.subdomain)
		}
	}
}

func Test_Dotify_Undotify_AddRemoveDoamin(t *testing.T) {
	domain := "test.tld"
	max := 151 - (len(domain) + 1)

	for i := 0; i < max; i++ {
		testcase, err := bc.GenerateRandomBytes(i)
		if err != nil {
			t.Error(err.Error())
		}

		dot, err := dns.Dotify(testcase)
		if err != nil {
			t.Error(err.Error())
		}
		fqdn := dns.AddDomain(domain, dot)
		subdomain := dns.RemoveDomain(domain, fqdn)

		if subdomain != dot {
			t.Error("Equality check failed:", subdomain, dot)
		}

		undot, err := dns.Undotify(subdomain)
		if err != nil {
			t.Error(err.Error())
		}

		if !bytes.Equal(testcase, undot) {
			t.Error("Equality check failed: ", testcase, len(testcase), undot, len(undot))
		}
	}
}
