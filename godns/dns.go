package main

import (
    "fmt"
    "net"
    "errors"
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

type Slash24Result struct {
    resolver string
    timeout_ips []string
    noname_ips []string
    ipToName map[string]string
} 

func rdns(lookupIP string, dnsServer string) (string, error) {
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
        return lookupIP, err
    }

    // TODO: add logic to store ips when the cnx times out
    msg := dns.Msg{}
    msg.SetQuestion(rev, dns.TypePTR)
    resp, _, err := c.Exchange(&msg, dnsServer+":53")
    if err != nil {
        // fmt.Println("ERROR", err, "ip: ",ip, "resolver: ", dnsServer)
        return "timeout", errors.New("query timeout")
    }
    if resp == nil {
        // fmt.Println(ip,"-", dnsServer,"failed PTR look up")
        return "a", nil
    }

    // TODO: condense code to single return statement
    // hostname := "None"
    if len(resp.Answer) > 0 {
        for _, ans := range resp.Answer {
            if ptr, ok := ans.(*dns.PTR); ok {
                hostname := ptr.Ptr
                // fmt.Println(dnsServer, ip, hostname)
                return hostname, nil


            } else {
                return "b", nil
            }
        }
    } else {
        // fmt.Println(lookupIP, "noname")
        return "noname", errors.New("query timeout")
    }

    return "c", nil

}

func lookUpSlash24(prefix string, dnsServer string) (Slash24Result) {
    /*
    The function will query th entire /24 of the prefix privded
    The entire lookup will take ~8s against 1.1.1.1 or 8.8.8.8
     Args:
        prefix (str): prefix of a /24 
        dnsServer (str): ip address of a dns server
    Returns:
        Nothing rn  
    */
    var noname_ips []string
    var timeout_ips []string

    var ipToName = make(map[string]string)
    for i := 0; i <= 255; i++ {
        ip := prefix + strconv.Itoa(i)
        // fmt.Println(ip)
        res, err := rdns(ip,dnsServer)
        if err != nil {
            if res == "noname" {
                noname_ips = append(noname_ips, ip)

            }
            if res == "timeout" {
                timeout_ips = append(timeout_ips, ip)
                // fmt.Println("timeout ", ip, dnsServer)

            }
        } else {
            ipToName[ip] = res
        }
        // fmt.Println(res)


        // fmt.Println(ip,res,"res")

    }


    // ipToName["resolver"] = dnsServer
    results := Slash24Result{dnsServer, timeout_ips, noname_ips, ipToName}
    // fmt.Println("results",results)
    // fmt.Println("names", len(ipToName))
    // fmt.Println("noname_ips", len(noname_ips))
    // fmt.Println("timeout_ips", len(timeout_ips))
    // ipToName["noname"] = noname_ips
    // ipToName["timeout"] = timeout_ips



    // fmt.Println(ipToName)
    // fmt.Println(ipToName["128.8.0.1"])
    return results
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
func process_results(c chan Slash24Result) {
// func process_results(c chan Slash24Result, wg sync.WaitGroup) {
    // defer wg.Done()
    fmt.Println("start")
    names := 0
    nonames := 0
    timeouts := 0
    for res := range c {
        timeouts += len(res.timeout_ips)
        nonames += len(res.noname_ips)
        names += len(res.ipToName)
        // fmt.Println("task result:", res)
    }
    fmt.Println("names:", names)
    fmt.Println("nonames:", nonames)
    fmt.Println("timeouts:", timeouts)
    fmt.Println("names/ip: ", names, "/", names+nonames+timeouts)
    fmt.Println("end")
}





func main() {
    /*

    */
    tasking := recv_tasking("/root/tasking.json")
    fmt.Println("tasking", tasking)
    var wg sync.WaitGroup
    c := make(chan Slash24Result)

    i := 0
    for _, task := range tasking {
        fmt.Println("task", task)
        wg.Add(1)

        go func(task Task, c chan Slash24Result) {
            defer wg.Done()
            // fmt.Println(task)
            result := lookUpSlash24(task.cidr, task.resolver)
            // go process_results(result)
            c <- result
            // fmt.Println(result)
        }(task, c)

        // fmt.Println(res)

        if i % 10 == 0 {
            time.Sleep(5 * time.Second)
        }

        i++
    }
    go func(c chan Slash24Result) {
        defer close(c)
        wg.Wait()
    }(c)

    process_results(c)

    // TODO: parse cnx details from input params
    // TODO: Connect to master node and listen for tasking
    // TODO: parse tasking message from master
    // TODO: dispatch tasking to work thread
    // TODO: send dns results to mysql db
    // TODO: test cnx to storage node
    // TODO: add rate limiting 



}
