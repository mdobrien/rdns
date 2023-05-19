package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    // "math"
    "net"
    "os"
    "strconv"
    "sync"
    "time"

    "github.com/cornelk/hashmap"
    "github.com/miekg/dns"
    "github.com/shogo82148/go-shuffle"
    "github.com/Workiva/go-datastructures/queue"
)

/*
At startup, a pool of ports is initialized. Then a port is grabbed from the pool 
to use for sending the dns request. If no ports are available, the dns request won't 
be sent until one is available.
*/

var ports = queue.New(1)
var desiredQPS = hashmap.New[string, int]()
var computedQPS = hashmap.New[string, int]()
var TASKING_PATH = "/tmp/tasking.json"
var RESOLVERS_PATH = "/tmp/resolvers.json"
var DATA_DIR = "/data/"
var QPS = 500

// var resolverToLatency = hashmap.New[string, int]()

// var ipTocount = hashmap.New[string, int]()
// var start = time.Now()
// qps = desiredQPS[resolver]
// numQueries = ipTocount[resolver]
// now = time.Now()
// delta = now.Sub(start)
// computedQPS = numQueries / sec(delta)

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
    prefix string
    latency float64
}

type ResolverStats struct {
    num_queries int
    num_timeouts int
    resolver string
}  


func init_logger() {
    // config logger
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)
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

    // This block will grab an available por from ports queue
    // Then bind to the port and send the dns request
    // get available port
    var port int
    for port == 0 {
        // log.Println(ports.Len(), "ports available")
        if ports.Len() % 1000 < 10 {
            log.Println("num ports available", ports.Len())
        }        
        if ports.Len() < 5000 {
            time.Sleep(10 * time.Second)
            log.Println("Waiting to have at least 60500 ports available")
            continue
        } else {
            val, err := ports.Get(1)
            if err != nil {
                // fmt.Println()
                return "ports", errors.New("failed to grab port number from the queue")
            }
            port = val[0].(int)
        }
    }
    c := new(dns.Client)
    laddr := net.UDPAddr{
        IP: net.ParseIP("[::1]"),
        Port: port,
        Zone: "",
    }
    c.Dialer = &net.Dialer{
        Timeout: 900 * time.Millisecond,
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
        return "timeout", errors.New("query timeout: ")
    }
    if resp == nil {
        // fmt.Println(ip,"-", dnsServer,"failed PTR look up")
        return "a", nil
    }
    // TODO: condense code to single return statement
    // hostname := "None"
    ports.Put(port)
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

func lookUpSlash24(prefix string, dnsServer string, wait int) (Slash24Result) {
    /*
    The function will query th entire /24 of the prefix privded
    If a dns queries times out. There will be a 15s pause before sending the next one
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

    if wait > 0 {
        time.Sleep(time.Duration(wait) * time.Millisecond)
    }

    st := time.Now()
    for i := 0; i <= 255; i++ {
        // time.Sleep(waitTime)

        ip := prefix + strconv.Itoa(i)
        // st := time.Now()
        // end := time.Now()
        // elapsed := end.Sub(st)
        // fmt.Println("resolver:", dnsServer, "ip:", ip, "time:", elapsed)
        res, err := rdns(ip,dnsServer)
        if err != nil {
            if res == "noname" {
                noname_ips = append(noname_ips, ip)
            }
            if res == "timeout" {
                timeout_ips = append(timeout_ips, ip)
                // log.Println("WARN timeoute: ip:", ip, " resolver: ", dnsServer)
                // sleep for 10 seconds
                // time.Sleep(1 * time.Second)
            }
        } else {
            ipToName[ip] = res
        }
    }

    end := time.Now()
    elapsed := end.Sub(st).Seconds()

    // fmt.Printf("%T %v\n", elapsed, elapsed)
    result := Slash24Result{dnsServer, timeout_ips, noname_ips, ipToName, prefix, elapsed}
    if len(timeout_ips) >= 100 {
        log.Println("resolver=", result.resolver, "prefix=", result.prefix + "0/24", "latency=", result.latency,  "timeouts=",len(timeout_ips), "nonames=",len(noname_ips), "names=",len(ipToName))
    }
    // fmt.Println("resolver:", dnsServer, "prefix:", prefix, "time:", elapsed )

    return result
}

func write_result(res Slash24Result) {
    _, err := os.Stat(DATA_DIR)
    if os.IsNotExist(err) {
        err := os.MkdirAll(DATA_DIR, 0750)
        if err != nil {
            log.Fatal(err)
        }
    }
    f, err := os.Create("/data/" + res.prefix)
    if err != nil {
        fmt.Println(err)
    }

    defer f.Close()
    json, _ := json.Marshal(res.ipToName)
    f.Write(json)
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
            tasks = append(tasks, task)
        }
    }

    shuffle.Slice(tasks)
    return tasks
}

func recv_resolvers(path string) []string {

    var resolvers []string

    // open file
    file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // create a  map to help the decoded data
    data := make(map[string][]string)

    //decode the jason data
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&data)
    if err != nil {
        log.Fatal(err)
    }

    // store data
    resolvers = data["resolvers"]

    return resolvers
}

func process_results(c chan Slash24Result) {
    stats := make(map[string]ResolverStats)
    total_names := 0
    total_nonames := 0
    total_timeouts := 0
    for res := range c {

        if res.resolver == "" {
            break
        }

        timeouts := len(res.timeout_ips)
        nonames := len(res.noname_ips)
        names := len(res.ipToName)
        total_names += names
        total_timeouts += timeouts
        total_nonames +=nonames

        log.Println("resolver=", res.resolver, "prefix=", res.prefix + "0/24", "latency=", res.latency,  "timeouts=",timeouts, "nonames=",nonames, "names=",names)
        val, ok := stats[res.resolver]
        if ok {
            val.num_queries += names+nonames+timeouts
            val.num_timeouts += timeouts
        } else {
            stats[res.resolver] = ResolverStats{names+nonames+timeouts, timeouts, res.resolver}
            // fmt.Println("task result:", res.resolver, "queries:", len(res.ipToName)+len(res.timeout_ips)+len(res.noname_ips),"names", len(res.ipToName),"timeouts:", len(res.timeout_ips), "nonames:", len(res.noname_ips))
        }
        // fmt.Println("task result:", res.resolver, "queries:", len(res.ipToName)+len(res.timeout_ips)+len(res.noname_ips),"names", len(res.ipToName),"timeouts:", len(res.timeout_ips), "nonames:", len(res.noname_ips))
        // fmt.Println(res.ipToName)
        // json, _ := json.Marshal(res.ipToName)
        // fmt.Println("json", string(json))
        write_result(res)

    }
    // log.Println("stats:", stats)
    log.Println("names/ip: ", total_names, "/", total_names+total_nonames+total_timeouts)
    log.Println("nonames/ip:", total_nonames, "/", total_names+total_nonames+total_timeouts)
    log.Println("timeouts/ip:", total_timeouts, "/", total_names+total_nonames+total_timeouts)
}

func rdns_worker(sendCH chan Slash24Result, recvCH chan Task, signalCH chan int, resolver string, wg sync.WaitGroup) {
    
    defer wg.Done()
    var lookup_wg sync.WaitGroup
    // var result Slash24Result
    // For some reason I do not understand updateInteval controls the size of the batches
    batchSize := 15
    batchSleep := 12
    updateInterval := 5
    resultCH := make(chan Slash24Result, 16777216)
    wait := 0

    i := 0
    // updateBatching := true
    for task := range recvCH {
        updateBatching := (i % updateInterval == 0)

        if task.resolver == "" {
        // if task.resolver == "" || i == 11 {
            // received terminate signal
            // break from recv on recvCH
            break
        }
            lookup_wg.Add(1)

        go func(updateBatching *bool, task Task, c chan Slash24Result, resultCH chan Slash24Result, wait int) {
            defer lookup_wg.Done()
            log.Println("processing: ", task)
            result := lookUpSlash24(task.cidr, task.resolver, wait)
            sendCH <- result

            log.Println("updateBatching:",*updateBatching, updateBatching)
            if *updateBatching {
                resultCH <- result
            }


        }(&updateBatching, task, sendCH, resultCH, wait)

        if updateBatching {
            result := <- resultCH    
            qps := float64(len(result.timeout_ips) + len(result.noname_ips) + len(result.ipToName)) / result.latency

            // batchSize 
            // batchSize := int(math.Round(float64(QPS) /qps))
            // batchSize := (QPS / int(qps)) + 1
            batchSize := 100
            // batchSleep
            // batchSleep := result.latency
            batchSleep := 100
            // updateInterval := 100
            updateBatching := false

            // Wait
            targetQPS, _ := desiredQPS.Get(result.resolver)
            wait := 1000 / targetQPS

            log.Println("updated values i=", i, ",elapsedTime =", result.latency, ", computed QPS=", qps, ", wait=", wait, ", batchSize=", batchSize, &batchSize,  ", batchSleep=", batchSleep, &batchSleep, 
               "updateBatching,", updateBatching, &updateBatching,  updateInterval)
        }

        if (i != 0) && (i % updateInterval == 0)  {
            log.Println("sleep for processing - resolver:", resolver, "batchSize: ", batchSize, &batchSize, "batchSleep:", batchSleep, &batchSleep)
            time.Sleep( time.Duration(batchSleep) * time.Second)
            log.Println(task.resolver, "completed", i, " tasks")
        }
        i++
    }
  
    // signal this worker finished processing tasking 
    lookup_wg.Wait()
    signalCH <- 1
    log.Println("Finished tasking for:", resolver)

}

func main() {

    init_logger()

    // set desired QPS for resolvers
    // TODO: parse desired qps from tasking req

    desiredQPS.Set("1.1.1.1", QPS)
    desiredQPS.Set("1.0.0.1", QPS)
    desiredQPS.Set("8.8.4.4", QPS)
    desiredQPS.Set("8.8.8.8", QPS)
    desiredQPS.Set("9.9.9.9", QPS)
    desiredQPS.Set("208.67.220.220", QPS)
    desiredQPS.Set("208.67.222.222", QPS)
    desiredQPS.Set("216.146.35.35",  QPS)
    desiredQPS.Set("74.82.42.42",  QPS)

    computedQPS.Set("1.1.1.1", 0)
    computedQPS.Set("1.0.0.1", 0)
    computedQPS.Set("8.8.4.4", 0)
    computedQPS.Set("8.8.8.8", 0)
    computedQPS.Set("208.67.220.220", 0)
    computedQPS.Set("208.67.222.222", 0)
    computedQPS.Set("216.146.35.35", 0)


    
    // init ports
    count := 1001
    for count <= 65500 {
        ports.Put(count)
        count++
    }

    // get list of resolvers
    resolvers := recv_resolvers(RESOLVERS_PATH)
    // fmt.Println("resolvers:",resolvers)

    // get tasking 
    tasking := recv_tasking(TASKING_PATH)
    log.Println("tasking:", tasking)
    // fmt.Println()

    // create wait group
    var wg sync.WaitGroup

    // create channel that will be used to pass results from
     // from workers to processing func
    resultCH := make(chan Slash24Result, 16777216)
    signalCH := make(chan int)
    
    // map that stores task channel for each resolver
    resolverToCH := make(map[string]chan Task)
    buffLen := len(tasking) + 1

    // setup channels for passing dns request tasking and
     // and init goroutines to listen for tasking
    // wg.Add(len(resolvers))
    for i := range resolvers {
        // increment wait group count for number of seolvers
        wg.Add(1)

        // create task channels and pair with resolver ip 
        resolver := resolvers[i]
        taskCH := make(chan Task, buffLen)
        resolverToCH[resolver] = taskCH

        // init goroutines to process tasking 
        go rdns_worker(resultCH, taskCH , signalCH, resolver , wg)
    }

    // dispatch tasks for specified resolver to correspodning gor
    for _, task := range tasking {
        
        log.Println("dispatched:", task)
        taskCH := resolverToCH[task.resolver]
        taskCH <- task
    }

    // send empty to signal all tasking has been sent
    for _,resolver := range resolvers {
        resolverToCH[resolver] <- Task{}
    }

    // listen for each gor to signal it finished processing tasks
    // then close close tasking buffered channel for all goroutines
    signals := 0
    for v := range signalCH {
        signals += v

        if signals == len(resolvers) {
            for _, taskCH := range resolverToCH {
                defer close(taskCH)
            }
            // close(resultCH)
            break
        }

    }

    resultCH <- Slash24Result{}
    // Write dns request results to disk
    process_results(resultCH)


    // TODO: parse cnx details from input params
    // TODO: Connect to master node and listen for tasking
    // TODO: parse tasking message from master
    // TODO: dispatch tasking to work thread
    // TODO: send dns results to mysql db
    // TODO: test cnx to storage node
    // TODO: add rate limiting 



}
