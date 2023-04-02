package main

import (
    "fmt"
    "net"
    "os"
    "log"
    "strconv"
    // "time"
    "github.com/miekg/dns"
    "sync"
    // "json"
    "encoding/json"
    "time"
    // "reflect"
)

type Task struct {
    resolver string
    cidr string
    qps int
}

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
    // c := dns.Client{}
    c := new(dns.Client)
    laddr := net.UDPAddr{
        IP: net.ParseIP("[::1]"),
        Port: 12345,
        Zone: "",
    }
    c.Dialer = &net.Dialer{
        Timeout: 200 * time.Millisecond,
        LocalAddr: &laddr,
    }


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
        fmt.Println("ERROR", err, "ip: ",ip, "resolver: ", dnsServer)
    }
    if resp == nil {
        // fmt.Println(ip,"-", dnsServer,"failed PTR look up")
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
    // for i := 0; i <= 255; i++ {
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

func recv_tasking(path string) []Task {
    // open file
    // read json
    // decode json
    // print json data
    var tasks []Task
    file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // create a  map to help the decoded data
    data := make(map[string][]string)

    // decode the json data
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&data)
    if err != nil {
        log.Fatal(err)
    }


    // Print the decoded data
    for key, values := range data {
        for _,cidr := range values {
            task := Task{key, cidr, 500}
            // fmt.Println(task)
            tasks = append(tasks, task)

        }
        // fmt.Printf("%s: %v\n", key, values)
    }

    // fmt.Println(tasks)
    // fmt.Println(tasks[0][0],tasks[0][1])
    // fmt.Println(reflect.TypeOf(tasks[0][1]))

    return tasks
}

func main() {
    /*

    */
    tasking := recv_tasking("/root/tasking.json")
    fmt.Println(tasking)
    var wg sync.WaitGroup

    i := 0
    for _, task := range tasking {
        // fmt.Println(task)
        wg.Add(1)

        go func(task Task) {
            defer wg.Done()
            // fmt.Println(task)
            lookUpSlash24(task.cidr, task.resolver)
        }(task)

        if i % 10 == 0 {
            time.Sleep(5 * time.Second)
        }

        i++
    }
    wg.Wait()

    // TODO: parse cnx details from input params
    // TODO: Connect to master node and listen for tasking
    // TODO: parse tasking message from master
    // TODO: dispatch tasking to work thread
    // TODO: send dns results to mysql db
    // TODO: test cnx to storage node
    // TODO: add rate limiting 



}
