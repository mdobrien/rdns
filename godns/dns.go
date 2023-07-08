package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "math"
    // "math/rand"
    "net"
    "os"
    "os/signal"
    "runtime"
    "strconv"
    "sync"
    "syscall"
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
var workerBatchSize = hashmap.New[string, int]()

var DATA_DIR = "/data/"
var LOG_FILE = "/root/debug.log"
var RESOLVERS_PATH = "/tmp/resolvers.json"
var TASKING_PATH = "/tmp/tasking.json"
var QPS = 500
var bs = 12

// var resolverToduration = hashmap.New[string, int]()

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
    duration float64
    timeStamp time.Time
}

type ResolverStats struct {
    num_queries int
    num_timeouts int
    resolver string
}  



func init_logger(file *os.File) {
    // config logger
    // Set output to write to both stdout and the file
    log.SetOutput(io.MultiWriter(os.Stdout, file))
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
        if ports.Len() % 1000 == 0 {
            numGoroutines := runtime.NumGoroutine()
            log.Println(dnsServer, "Number of running Goroutines:", numGoroutines)
        }        
        if ports.Len() < 5000 {
            // log.Println(dnsServer, "Waiting to have at least 50000 ports available")
            for ports.Len() < 65000 {
                time.Sleep(10 * time.Second)

                log.Println(dnsServer, "waiting for ports:", ports.Len())
                numGoroutines := runtime.NumGoroutine()
                log.Println(dnsServer, "Number of running Goroutines:", numGoroutines)


            }
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
        ports.Put(port)
        return "timeout", errors.New("query timeout: ")
    }
    if resp == nil {
        // fmt.Println(ip,"-", dnsServer,"failed PTR look up")
        ports.Put(port)
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
    result := Slash24Result{dnsServer, timeout_ips, noname_ips, ipToName, prefix, elapsed, time.Now()}
    if len(timeout_ips) >= 100  {
        log.Println("resolver=", result.resolver, "prefix=", result.prefix + "0/24", "duration=", result.duration,  "timeouts=",len(timeout_ips), "nonames=",len(noname_ips), "names=",len(ipToName))
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

        log.Println("resolver=", res.resolver, "prefix=", res.prefix + "0/24", "duration=", res.duration,  "timeouts=",timeouts, "nonames=",nonames, "names=",names, "ts=", res.timeStamp.Format("2006-01-02 15:04:05.000"), res.timeStamp.UnixMilli())
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
    // batchSize := 2
    batchSize, _ := workerBatchSize.Get(resolver)
    _ = batchSize
    // batchSleep := 12
    updateInterval := 6
    resultCH := make(chan Slash24Result, 16777216)
    // intervalCH := make(chan int, batchSize)
    wait := 0
    log.Println("stared worker for", resolver)
    i := 1
    count := 0
    _ = count


    // updateBatching := true
    for task := range recvCH {
        updateBatching := (i % updateInterval == 0)
        numGoroutines := runtime.NumGoroutine()

        if ports.Len() < 10000 {
            for numGoroutines > 15  {
                numGoroutines := runtime.NumGoroutine()
                _ = numGoroutines
                time.Sleep(2 * time.Second)
                log.Println(resolver, "Number of running Goroutines:", numGoroutines, "ports:",  ports.Len() )

            }
        }


        if i % batchSize == 0 || task.resolver == "" {
            lookup_wg.Wait()
            log.Println("num ports available", ports.Len())
            log.Println(resolver, "Number of running Goroutines:", numGoroutines)
            // buf := make([]byte, 1024)
            // n := runtime.Stack(buf, false)
            // fmt.Printf("Stack trace:\n%s\n", buf[:n])            

            // finished processing tasking
            if task.resolver == "" {
                break
            }
        }
        task.resolver = resolver

        lookup_wg.Add(1)
        go func(updateBatching *bool, task Task, c chan Slash24Result, resultCH chan Slash24Result, wait int) {
            defer lookup_wg.Done()
            task.resolver = resolver
            log.Println("processing: ", task.cidr + "0/24")
            result := lookUpSlash24(task.cidr, task.resolver, wait)

            sendCH <- result
            // randomNumber := rand.Intn(5) + 1
            // time.Sleep(time.Duration(randomNumber) * time.Second)
            // time.Sleep(100 * time.Millisecond)
            // log.Println("updateBatching:",*updateBatching, updateBatching)
            qps := float64(len(result.timeout_ips) + len(result.noname_ips) + len(result.ipToName)) / result.duration
            // log.Println(resolver,  "to=" ,len(result.timeout_ips), "nn=",len(result.noname_ips),"names=",len(result.ipToName), "task duration=", result.duration, ", computed QPS=", qps, "ts=", result.timeStamp.Format("2006-01-02 15:04:05.000"))

            threshold := 3.14

            epsilon := 0.000001 // Define an acceptable tolerance for equality

            if math.Abs(threshold-result.duration) < epsilon {
                // fmt.Println("threshold is approximately equal to duration")
            } else if threshold < result.duration {
                //
            } else {
                log.Println("WARN", result, "being rate limited", task)
            }
            
            log.Println(
                "Finished: ",
                resolver,
                task.cidr + "0/24",
                "to=" ,len(result.timeout_ips),
                "nn=",len(result.noname_ips),
                "names=",len(result.ipToName),
                "task duration=", result.duration,
                "computed QPS=", qps,
                "ts=", result.timeStamp.Format("2006-01-02 15:04:05.000"), result.timeStamp.UnixMilli())                        

            // log.Println("Finished: ", task, result.timeStamp)

        }(&updateBatching, task, sendCH, resultCH, wait)

        // if updateBatching {
        //     batchSize++
        // }
        // if updateBatching {
        //     result := <- resultCH    
        //     qps := float64(len(result.timeout_ips) + len(result.noname_ips) + len(result.ipToName)) / result.duration

        //     // batchSize 
        //     // batchSize := int(math.Round(float64(QPS) /qps))
        //     // batchSize := (QPS / int(qps)) + 1
        //     batchSize += 2
        //     // batchSleep
        //     // batchSleep := result.duration
        //     batchSleep := 12
        //     // updateInterval := 100
        //     updateBatching := false

        //     // Wait
        //     targetQPS, _ := desiredQPS.Get(result.resolver)
        //     wait := 1000 / targetQPS

        //     log.Println(resolver, "updated values i=", i, ",elapsedTime =", result.duration, ", computed QPS=", qps, ", wait=", wait, ", batchSize=", batchSize, &batchSize,  ", batchSleep=", batchSleep, &batchSleep, 
        //        "updateBatching,", updateBatching, &updateBatching,  updateInterval)
        // }


        i++
    }
  
    // signal this worker finished processing tasking 
    lookup_wg.Wait()
    signalCH <- 1
    log.Println("Finished tasking for:", resolver)

}


// func rdns_worker(sendCH chan Slash24Result, recvCH chan Task, signalCH chan int, resolver string, wg sync.WaitGroup) {
    
//     defer wg.Done()
//     var lookup_wg sync.WaitGroup
//     // var result Slash24Result
//     // For some reason I do not understand updateInteval controls the size of the batches
//     batchSize := 2
//     // batchSleep := 12
//     updateInterval := 4
//     resultCH := make(chan Slash24Result, 16777216)
//     intervalCH := make(chan int, 1000000)
//     wait := 0
//     log.Println("stared worker for", resolver)
//     i := 0
//     count := 0
//     _ = count


//     // updateBatching := true
//     for task := range recvCH {
//         updateBatching := (i % updateInterval == 0)

//         if task.resolver == "" {
//         // if task.resolver == "" || i == 11 {
//             // received terminate signal
//             // break from recv on recvCH
//             intervalCH <- 0
//             break
//         }
//             lookup_wg.Add(1)

//         go func(updateBatching *bool, task Task, intervalCH chan int, c chan Slash24Result, resultCH chan Slash24Result, wait int) {
//             defer lookup_wg.Done()
//             task.resolver = resolver
//             log.Println("processing: ", task)
//             result := lookUpSlash24(task.cidr, task.resolver, wait)

//             sendCH <- result
//             // randomNumber := rand.Intn(5) + 1
//             // time.Sleep(time.Duration(randomNumber) * time.Second)
//             time.Sleep(5 * time.Second)
//             // log.Println("updateBatching:",*updateBatching, updateBatching)
//             if *updateBatching {
//                 // resultCH <- result
//             }
//             intervalCH <- 1
//             log.Println("Finished: ", task, result.timeStamp)

//         }(&updateBatching, task, intervalCH, sendCH, resultCH, wait)

//         // if updateBatching {
//         //     result := <- resultCH    
//         //     qps := float64(len(result.timeout_ips) + len(result.noname_ips) + len(result.ipToName)) / result.duration

//         //     // batchSize 
//         //     // batchSize := int(math.Round(float64(QPS) /qps))
//         //     // batchSize := (QPS / int(qps)) + 1
//             // batchSize += 2
//         //     // batchSleep
//         //     // batchSleep := result.duration
//         //     batchSleep := 12
//         //     // updateInterval := 100
//         //     updateBatching := false

//         //     // Wait
//         //     targetQPS, _ := desiredQPS.Get(result.resolver)
//         //     wait := 1000 / targetQPS

//         //     log.Println(resolver, "updated values i=", i, ",elapsedTime =", result.duration, ", computed QPS=", qps, ", wait=", wait, ", batchSize=", batchSize, &batchSize,  ", batchSleep=", batchSleep, &batchSleep, 
//         //        "updateBatching,", updateBatching, &updateBatching,  updateInterval)
//         // }

//         if (i != 0) && (i % batchSize == 0)  {
//             for v := range intervalCH {
//                 log.Println(task.cidr, "i=",i, "v=", v, "count=",count, "batchSize=", batchSize)
//                 if count < batchSize - 1   || v == 0 {
//                     count := 0
//                     _ = count
//                     // time.Sleep(5 * time.Second)
//                     break
//                 }
//                 count += v
//                 _ = count
//                 // log.Println(count)
//             }
//             log.Println(task.resolver, "completed", i, " tasks")
//         }
//         i++
//     }
  
//     // signal this worker finished processing tasking 
//     lookup_wg.Wait()
//     signalCH <- 1
//     log.Println("Finished tasking for:", resolver)

// }

func catchSignal() {

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    // Wait for the termination signal
    sig := <-c

    // Convert the signal to its corresponding name
    signalName := sig.String()

    // Display the signal name
    fmt.Printf("Received termination signal: %s\n", signalName)

}

func main() {

    file, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }

    init_logger(file)
    defer file.Close()

    go catchSignal()

    // set desired QPS for resolvers
    // TODO: parse desired qps from tasking req

    desiredQPS.Set("1.1.1.1", QPS)
    desiredQPS.Set("1.0.0.1", QPS)
    desiredQPS.Set("8.8.4.4", QPS)
    desiredQPS.Set("8.8.8.8", QPS)
    desiredQPS.Set("9.9.9.9", QPS)
    desiredQPS.Set("114.114.114.114", QPS)
    desiredQPS.Set("114.114.115.115", QPS)
    desiredQPS.Set("208.67.220.220", QPS)
    desiredQPS.Set("208.67.222.222", QPS)
    desiredQPS.Set("216.146.35.35",  QPS)
    desiredQPS.Set("74.82.42.42",  QPS)
    desiredQPS.Set("4.2.2.1",  QPS)
    desiredQPS.Set("156.154.70.1",  QPS)
    desiredQPS.Set("8.20.247.20",  QPS)
    desiredQPS.Set("185.228.169.9",  QPS)
    desiredQPS.Set("199.85.126.10",  QPS)
    desiredQPS.Set("84.200.69.80",  QPS)
    desiredQPS.Set("64.6.64.6",  QPS)
    desiredQPS.Set("199.7.91.13",  QPS)
    desiredQPS.Set("199.7.83.42",  QPS)
    desiredQPS.Set("80.80.80.80",  QPS)
    desiredQPS.Set("156.154.71.1",  QPS)
    desiredQPS.Set("84.200.70.40",  QPS)
    desiredQPS.Set("209.244.0.3",  QPS)
    desiredQPS.Set("209.244.0.3", QPS)
    desiredQPS.Set("199.85.126.30", QPS)
    desiredQPS.Set("199.85.127.20", QPS)
    desiredQPS.Set("94.140.14.14", QPS)
    desiredQPS.Set("94.140.14.140", QPS)
    desiredQPS.Set("94.140.14.141", QPS)
    desiredQPS.Set("94.140.14.15", QPS)
    desiredQPS.Set("94.140.15.15", QPS)
    desiredQPS.Set("94.140.15.16", QPS)
    desiredQPS.Set("64.6.65.6", QPS)
    desiredQPS.Set("185.228.168.9", QPS)
    desiredQPS.Set("74.82.42.42", QPS)
    desiredQPS.Set("8.20.247.20", QPS)
    desiredQPS.Set("209.244.0.4", QPS)
    desiredQPS.Set("4.2.2.2", QPS)
    desiredQPS.Set("149.112.112.10", QPS)
    desiredQPS.Set("45.90.28.0", QPS)
    desiredQPS.Set("45.90.30.0", QPS)
    desiredQPS.Set("195.46.39.39", QPS)
    desiredQPS.Set("156.154.71.1", QPS)
    desiredQPS.Set("199.85.126.20", QPS)
    desiredQPS.Set("199.85.127.10", QPS)
    desiredQPS.Set("199.85.127.30", QPS)
    desiredQPS.Set("176.103.130.130", QPS)
    desiredQPS.Set("8.26.56.26", QPS)
    desiredQPS.Set("149.112.122.10", QPS)
    desiredQPS.Set("149.112.112.112", QPS)
    desiredQPS.Set("149.112.121.10", QPS)
    desiredQPS.Set("9.9.9.10", QPS)
    desiredQPS.Set("185.228.168.10", QPS)
    desiredQPS.Set("176.103.130.131", QPS)
    desiredQPS.Set("75.75.76.76", QPS)
    desiredQPS.Set("75.75.75.75", QPS)
    desiredQPS.Set("8.8.8.8", QPS)
    desiredQPS.Set("195.46.39.40", QPS)
    desiredQPS.Set("4.2.2.1", QPS)
    desiredQPS.Set("185.228.168.168", QPS)
    desiredQPS.Set("1.1.1.1", QPS)
    desiredQPS.Set("193.17.47.1", QPS)
    desiredQPS.Set("185.228.169.11", QPS)
    desiredQPS.Set("185.228.169.9", QPS)
    desiredQPS.Set("185.43.135.1", QPS)
    desiredQPS.Set("45.11.45.11", QPS)
    desiredQPS.Set("80.67.169.12", QPS)
    desiredQPS.Set("80.67.169.40", QPS)
    desiredQPS.Set("80.80.80.80", QPS)
    desiredQPS.Set("80.80.81.81", QPS)
    desiredQPS.Set("77.88.8.1", QPS)
    desiredQPS.Set("77.88.8.3", QPS)
    desiredQPS.Set("77.88.8.7", QPS)
    desiredQPS.Set("77.88.8.8", QPS)
    desiredQPS.Set("1.2.3.4", QPS)

    workerBatchSize.Set("1.2.3.4", bs)
    workerBatchSize.Set("114.114.114.114", 5)
    workerBatchSize.Set("114.114.115.115", 5)
    workerBatchSize.Set("64.6.64.6",  3)
    workerBatchSize.Set("9.9.9.9", 3)
    workerBatchSize.Set("156.154.70.1", 3)
    workerBatchSize.Set("199.85.126.10",  3)
    workerBatchSize.Set("1.1.1.1", bs)
    workerBatchSize.Set("1.0.0.1", bs)
    workerBatchSize.Set("8.8.4.4", bs)
    workerBatchSize.Set("8.8.8.8", bs)
    workerBatchSize.Set("208.67.220.220", bs)
    workerBatchSize.Set("208.67.222.222", 3)
    workerBatchSize.Set("216.146.35.35",  bs)
    workerBatchSize.Set("74.82.42.42",  bs)
    workerBatchSize.Set("4.2.2.1",  bs)
    workerBatchSize.Set("8.20.247.20",  bs)
    workerBatchSize.Set("185.228.169.9",  bs)
    workerBatchSize.Set("84.200.69.80",  bs)
    workerBatchSize.Set("199.7.91.13",  bs)
    workerBatchSize.Set("199.7.83.42",  bs)
    workerBatchSize.Set("80.80.80.80",  bs)
    workerBatchSize.Set("156.154.71.1",  bs)
    workerBatchSize.Set("84.200.70.40",  bs)
    workerBatchSize.Set("209.244.0.3",  bs)
    workerBatchSize.Set("209.244.0.3", bs)
    workerBatchSize.Set("199.85.126.30", bs)
    workerBatchSize.Set("199.85.127.20", bs)
    workerBatchSize.Set("94.140.14.14", bs)
    workerBatchSize.Set("94.140.14.140", bs)
    workerBatchSize.Set("94.140.14.141", bs)
    workerBatchSize.Set("94.140.14.15", bs)
    workerBatchSize.Set("94.140.15.15", bs)
    workerBatchSize.Set("94.140.15.16", bs)
    workerBatchSize.Set("64.6.65.6", bs)
    workerBatchSize.Set("185.228.168.9", bs)
    workerBatchSize.Set("74.82.42.42", bs)
    workerBatchSize.Set("8.20.247.20", bs)
    workerBatchSize.Set("209.244.0.4", bs)
    workerBatchSize.Set("4.2.2.2", bs)
    workerBatchSize.Set("149.112.112.10", bs)
    workerBatchSize.Set("45.90.28.0", bs)
    workerBatchSize.Set("45.90.30.0", bs)
    workerBatchSize.Set("195.46.39.39", bs)
    workerBatchSize.Set("156.154.71.1", bs)
    workerBatchSize.Set("199.85.126.20", bs)
    workerBatchSize.Set("199.85.127.10", bs)
    workerBatchSize.Set("199.85.127.30", bs)
    workerBatchSize.Set("176.103.130.130", bs)
    workerBatchSize.Set("8.26.56.26", bs)
    workerBatchSize.Set("149.112.122.10", bs)
    workerBatchSize.Set("149.112.112.112", bs)
    workerBatchSize.Set("149.112.121.10", bs)
    workerBatchSize.Set("9.9.9.10", bs)
    workerBatchSize.Set("185.228.168.10", bs)
    workerBatchSize.Set("176.103.130.131", bs)
    workerBatchSize.Set("75.75.76.76", bs)
    workerBatchSize.Set("75.75.75.75", bs)
    workerBatchSize.Set("8.8.8.8", bs)
    workerBatchSize.Set("195.46.39.40", bs)
    workerBatchSize.Set("4.2.2.1", bs)
    workerBatchSize.Set("185.228.168.168", bs)
    workerBatchSize.Set("1.1.1.1", bs)
    workerBatchSize.Set("193.17.47.1", bs)
    workerBatchSize.Set("185.228.169.11", bs)
    workerBatchSize.Set("185.228.169.9", bs)
    workerBatchSize.Set("185.43.135.1", bs)
    workerBatchSize.Set("45.11.45.11", bs)
    workerBatchSize.Set("80.67.169.12", bs)
    workerBatchSize.Set("80.67.169.40", bs)
    workerBatchSize.Set("80.80.80.80", bs)
    workerBatchSize.Set("80.80.81.81", bs)
    workerBatchSize.Set("77.88.8.1", bs)
    workerBatchSize.Set("77.88.8.3", bs)
    workerBatchSize.Set("77.88.8.7", bs)
    workerBatchSize.Set("77.88.8.8", bs)
    

    
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
    log.Println("tasking:", len(tasking), time.Now().UnixMilli())
    // fmt.Println()

    // create wait group
    var wg sync.WaitGroup

    buffLen := len(tasking) + 1
    // create channel that will be used to pass results from
     // from workers to processing func
    resultCH := make(chan Slash24Result, 16777216)
    signalCH := make(chan int)
    taskCH := make(chan Task, buffLen)
    // map that stores task channel for each resolver
    // resolverToCH := make(map[string]chan Task)

    // setup channels for passing dns request tasking and
     // and init goroutines to listen for tasking
    // wg.Add(len(resolvers))
    for i := range resolvers {
        // increment wait group count for number of seolvers
        wg.Add(1)

        // create task channels and pair with resolver ip 
        resolver := resolvers[i]
        // taskCH := make(chan Task, buffLen)
        // resolverToCH[resolver] = taskCH
        time.Sleep(1 * time.Second)

        // init goroutines to process tasking 
        go rdns_worker(resultCH, taskCH , signalCH, resolver , wg)
    }

    // dispatch tasks for specified resolver to correspodning gor
    for _, task := range tasking {
        
        log.Println("dispatched:", task)
        // taskCH := resolverToCH[task.resolver]
        taskCH <- task
    }

    // send empty to signal all tasking has been sent
    for range resolvers {
        taskCH <- Task{}
        // resolverToCH[resolver] <- Task{}
    }

    // listen for each gor to signal it finished processing tasks
    // then close close tasking buffered channel for all goroutines
    signals := 0
    for v := range signalCH {
        signals += v

        if signals == len(resolvers) {
            // for _, taskCH := range resolverToCH {
            defer close(taskCH)
            // }
            // close(resultCH)
            break
        }

    }

    resultCH <- Slash24Result{}
    // Write dns request results to disk
    process_results(resultCH)
    log.Println("Made it to here")


    // TODO: parse cnx details from input params
    // TODO: Connect to master node and listen for tasking
    // TODO: parse tasking message from master
    // TODO: dispatch tasking to work thread
    // TODO: send dns results to mysql db
    // TODO: test cnx to storage node
    // TODO: add rate limiting 
}

