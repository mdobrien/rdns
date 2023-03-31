package main

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
    /*

    Perform reverse DNS lookup on lookupIP and execute the query against dnsServer 
    Args:
        lookupIP (str) : ip address to conduct PTR DNS request
        dnsServer (str): ip address of DNS server queries
    */
    // print params
    // fmt.Println(lookupIP, dnsServer)

    // create client
    c := dns.Client{}


    ip := net.ParseIP(lookupIP)  // Todo: check if this is redundent
    rev, err := dns.ReverseAddr(ip.String())
    if err != nil {
        fmt.Println(err)
    }

    // TODO: add logic to store ips when the cnx times out
    msg := dns.Msg{}
    msg.SetQuestion(rev, dns.TypePTR)
    resp, _, err := c.Exchange(&msg, dnsServer+":53")
    if err != nil {
        fmt.Println(err)
    }
    if resp == nil {
        fmt.Println(ip,"failed PTR look up")
        return ""
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

func lookUpSlash24(prefix string, dnsServer string) (map[string]string) {
    /*
    The function will query th entire /24 of the prefix privded
    The entire lookup will take ~8s against 1.1.1.1 or 8.8.8.8
     Args:
        prefix (str): prefix of a /24 
        dnsServer (str): ip address of a dns server
    Returns:
        Nothing rn  
    */
    var ipToName = make(map[string]string)
    for i := 0; i <= 255; i++ {
        ip := prefix + strconv.Itoa(i)
        // fmt.Println(ip)
        res := rdns(ip,dnsServer)
        if len(res) > 0 {
            ipToName[ip] = res
        }


        // fmt.Println(ip,res,"res")

    }

    fmt.Println(ipToName)
    // fmt.Println(ipToName["128.8.0.1"])
    return ipToName
}

func main() {
    /*

    */

    // TODO: parse cnx details from input params
    // TODO: Connect to master node and listen for tasking
    // TODO: parse tasking message from master
    // TODO: dispatch tasking to work thread
    // TODO: send dns results to mysql db

    // TODO: test cnx to storage node
    // TODO: add rate limiting 
    var wg sync.WaitGroup
    var prefixes = []string{
        "128.8.0.",
        // "128.8.1.",
        // "128.8.2.",
        // "128.8.3.",
        // "128.8.4.",
        // "128.8.5.",
        // "128.8.6.",
        // "128.8.7.",
        // "128.8.8.",
        // "128.8.9.",
    }
    for _, prefix := range prefixes {
        wg.Add(1)

        go func(prefix string) {
            defer wg.Done()
            lookUpSlash24(prefix, "8.8.4.4")
            // ipToName := lookUpSlash24(prefix, "8.8.4.4")
            // TODO:  sendToDB(ipToName)
        }(prefix)
    }
    wg.Wait()




}
