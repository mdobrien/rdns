package main



/*
    https://pkg.go.dev/github.com/miekg/dns#PTR

*/
import (
    "fmt"
    "net"
    // "os"

    "github.com/miekg/dns"
)

func main() {

    c := dns.Client{}
    ip := net.ParseIP("1.1.1.1")
    rev, err := dns.ReverseAddr(ip.String())
    if err != nil {
        // handle error
    }
    msg := dns.Msg{}
    msg.SetQuestion(rev, dns.TypePTR)
    resp, _, err := c.Exchange(&msg, "8.8.8.8:53")
    if err != nil {
        fmt.Println(err)
        // handle error
    }

    if len(resp.Answer) > 0 {
        for _, ans := range resp.Answer {
            if ptr, ok := ans.(*dns.PTR); ok {
                hostname := ptr.Ptr
                fmt.Println(hostname)
                // do something with the hostname
            }
        }
    }
    fmt.Println("hello")




}
