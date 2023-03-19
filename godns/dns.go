package main

import (
    "fmt"
    "net"
    "os"

    "github.com/miekg/dns"
)

func main() {
    // check command-line arguments
    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s <ip address> <dns server>\n", os.Args[0])
        os.Exit(1)
    }

    ip := os.Args[1]
    dnsServer := os.Args[2]

    // create DNS message
    msg := new(dns.Msg)
    msg.SetQuestion(dns.Fqdn(ip), dns.TypePTR)

    // create DNS client
    client := new(dns.Client)
    server := net.ParseIP(dnsServer)
    if server == nil {
        fmt.Fprintf(os.Stderr, "Invalid DNS server: %s\n", dnsServer)
        os.Exit(1)
    }
    // dnsTimeout := 1
    client.Net = "udp"
    client.ReadTimeout = 2 
    client.WriteTimeout = 2 

    // send DNS query
    resp, _, err := client.Exchange(msg, server.String()+":53")
    if err != nil {
        fmt.Fprintf(os.Stderr, "DNS query error: %s\n", err)
        os.Exit(1)
    }

    // print DNS response
    if resp.Rcode != dns.RcodeSuccess {
        fmt.Fprintf(os.Stderr, "Non-successful DNS response: %s\n", dns.RcodeToString[resp.Rcode])
        os.Exit(1)
    }

    for _, answer := range resp.Answer {
        if ptr, ok := answer.(*dns.PTR); ok {
            fmt.Println(ptr.Ptr)
        }
    }
}
