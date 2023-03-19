package main



/*
    https://pkg.go.dev/github.com/miekg/dns#PTR

*/
import (
    "fmt"
    "net"
    // "os"
    "strconv"
    // "time"
    "github.com/miekg/dns"
    "sync"
)

func rdns(lookupIP string, dnsServer string) (string) {
    
    // print params
    // fmt.Println(lookupIP, dnsServer)

    // create client
    c := dns.Client{}


    ip := net.ParseIP(lookupIP)  // Todo: check if this is redundent
    rev, err := dns.ReverseAddr(ip.String())
    if err != nil {
        fmt.Println(err)
    }
    msg := dns.Msg{}
    msg.SetQuestion(rev, dns.TypePTR)
    resp, _, err := c.Exchange(&msg, dnsServer+":53")
    if err != nil {
        fmt.Println(err)
    }

    // TODO: condense code to single return statement
    // hostname := "None"
    if len(resp.Answer) > 0 {
        for _, ans := range resp.Answer {
            if ptr, ok := ans.(*dns.PTR); ok {
                hostname := ptr.Ptr
                // fmt.Println(dnsServer, ip, hostname)
                return hostname


            } else {
                return ""
            }
        }
    } else {
        return ""
    }
    return ""

}

func lookUpSlash24(prefix string, dnsServer string) {
    /*
    The function will query th entire /24 of the prefix privded
    The entire lookup will take ~8s against 1.1.1.1 or 8.8.8.8
     Args:
        prefix (str): prefix of a /24 
        dnsServer (str): ip address of a dns server
    Returns:
        Nothing rn
     do a full rdns of /24    
    */
    var ipToName = make(map[string]string)
    for i := 0; i <= 255; i++ {
        ip := prefix + strconv.Itoa(i)
        // fmt.Println(ip)
        res := rdns(ip,dnsServer)
        // fmt.Println(ip,res,"res")

        ipToName[ip] = res
    }

    fmt.Println(ipToName)
    // fmt.Println(ipToName["128.8.0.1"])
}

func main() {
    /*

    */

    var wg sync.WaitGroup
    var prefixes = []string{
        // "128.2.0.",
        "128.8.0.",
        "128.8.1.",
        "128.8.2.",
        "128.8.3.",
        "128.8.4.",
        "128.8.5.",
        "128.8.6.",
        "128.8.7.",
        "128.8.8.",
        "128.8.9.",
        "128.8.10.",
        "128.8.11.",
        "128.8.12.",
        "128.8.13.",
        "128.8.14.",
        "128.8.15.",
        "128.8.16.",
        "128.8.17.",
    }

    for _, prefix := range prefixes {
        wg.Add(1)

        go func(prefix string) {
            defer wg.Done()
            lookUpSlash24(prefix, "8.8.4.4")
            // lookUpSlash24(prefix, "1.1.1.1")
        }(prefix)
    }
    wg.Wait()




}
