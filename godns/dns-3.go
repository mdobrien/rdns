package main

import (
    "fmt"
    // "net"
    "os"
    // "time"

    "github.com/miekg/dns"
)

// Change query from MX lookup to PTR
// parse output	



func main() {
    // check command-line arguments
    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s <ip address> <dns server>\n", os.Args[0])
        os.Exit(1)
    }

    // ip := os.Args[1]
    ip := os.Args[1]
    dnsServer := os.Args[2]
    fmt.Println(ip)
    fmt.Println(dnsServer)


    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn("miek.nl"), dns.TypeMX)
    m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{"miek.nl.", dns.TypeMX, dns.ClassINET}

	c := new(dns.Client)

	// TODO:change 8.8.8.8 to use dnsServer param
	in, rtt, err := c.Exchange(m1, "8.8.8.8:53")

	fmt.Println(in)
	fmt.Println(rtt)
	fmt.Println(err)

	c.SingleInflight = true // suppressin multiple outstanding queries



}
