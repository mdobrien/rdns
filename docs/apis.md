apis.md



python3 rdns.py


python3 rdns.py run --cidr 128.0.0.0/8 --destination 1.1.1.1 --qps 500 

Variables
  - number of workers
  - Number of DNS servers
  - qps/qpm threshold per DNS server (think cloudflare)
  - qps/qpm threshold per authoritative DNS server 
  - # number of ips/cidrs to do rdns
  - desire time to completion


Requirements
  - distrubute lookups across N workers
  - For each worker, partition work so worker is using N DNS servers
  - workers should pass data to an aggregator node
  - workers use dns.go for looks. workers use rdns.py for ochestration
  - partition work equally among workers
  - workers will distribute looks ups evenly across available name servers
  - work will back up on queries if being rate limitd(store stats resolve or autoritative name server rate limit, QPS and sustained qps against both targets)
  - workers communicate with master via shared named pipe
  - show qpm,qps, aggegrate qpm, num workers, # dns servers, and runtime (hours/mins) before starting


python3 rdns.py run --cidr 128.8.0.0/16 129.2.0.0/16  --resolvers 1.1.1.1 8.8.8.8 8.8.4.4 --qps 500 --workers 2

# Work breakdown
  - convert list of CIDRS into a list of /24s to be lookuped
  - distribute /24s evenly accross workers

Worker class
  - init(nameServers, cidrs, qps)
  - 

# DNS query stats
aggregate qpm	300000
qpm	30000
qps	500
#DNS servers	10
#workers	1
runtime (minutes)	55.92405333
runtime (hours)	0.9320675556