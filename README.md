
02:15:49 == 10:16 PM
# rdns
	* completed a /14 (524288 ips) in 1523s (25.38m)
	* 25.37*((2^24)/524288)


# Quick start
```bash
docker run -ti --volume $RDNS_HOME/rdns.py:/root/rdns.py --volume $RDNS_HOME/godns:/root/godns rdns bash

# in the docker container 
# step 1
cd /root/godns
# step 2
go get github.com/Workiva/go-datastructures && \
go get github.com/miekg/dns && \
go get github.com/cornelk/hashmap && \
go get github.com/shogo82148/go-shuffle && \
go get golang.org/x/net && \
go get golang.org/x/sys
# step 3
python3 ../rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500
```

# Setup
```bash
# Set project env vars and aliases
export RDNS_HOME='/path/to/<rdns>'
alias db='cd $RDNS_HOME'
alias biz='docker run -ti --volume $RDNS_HOME/rdns.py:/root/rdns.py --volume $RDNS_HOME/godns:/root/godns rdns bash'

# Build container
docker build -t rdns .

# run container
# docker run -ti --volume $RDNS_HOME/rdns.py:/root/rdns.py --volume $RDNS_HOME/godns:/root/godns rdns bash
docker run \
	-ti \
	--volume $RDNS_HOME/rdns.py:/rdns/src/rdns.py \
	--volume  $RDNS_HOME/godns/dns.go:/rdns/src/dns.go \
	rdns \
	/bin/bash

# Next goal
rdns.py run  --cidr 128.0.0.0/8 --resolvers 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 9.9.9.9 --qps 500
rdns.py run --cidr 128.0.0.0/9 --resolvers 64.6.64.6 156.154.70.1  199.85.126.10  199.85.127.20 1.1.1.1  1.0.0.1  8.8.8.8  8.8.4.4  208.67.222.222  208.67.220.220 209.244.0.3 209.244.0.4 74.82.42.42 149.112.122.10  --qps 500

# Goat resolver combo this far
rdns.py run --cidr 128.0.0.0/10 --resolvers 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 --qps 500
# Passed goal
rdns.py run --cidr 128.0.0.0/13 --resolvers 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 --qps 500
rb && rdns.py run --cidr     128.0.0.0/14 --resolvers 1.1.1.1 8.8.8.8 8.8.4.4 1.0.0.1 74.82.42.42 208.67.222.222 4.2.2.1 --qps 500
rb && rdns.py run --cidr     66.0.0.0/15 --resolvers 1.1.1.1 8.8.8.8 8.8.4.4 1.0.0.1 74.82.42.42 208.67.222.222 4.2.2.1 --qps 500
# example run
rdns.py run --cidr 128.8.0.0/23 --resolvers 1.1.1.1 8.8.8.8 --qps 500
rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500
rdns.py run --cidr     66.30.0.0/16 --resolvers 1.1.1.1 1.0.0.1 8.8.4.4 8.8.8.8 --qps 500
rdns.py run --cidr 128.8.0.0/16 --qps 300 --resolvers 1.1.1.1 8.8.8.8 9.9.9.9 208.67.222.222 64.6.64.6 4.2.2.1 8.26.56.26 84.200.69.80 77.88.8.8 185.228.168.9 156.154.70.1 199.85.126.10 195.46.39.39 74.82.42.42

```


# Test Networks
	128.2.0.0/16 	Carnegie Mellon University
	128.237.0.0/16  Carnegie Mellon University
	128.8.0.0/16 	University of Maryland
	129.2.0.0/16 	University of Maryland 
	147.222.0.0/16 	Gonzaga University 
	140.247.0.0/16 	Harvard University

# Test 1
	python3 ../rdns.py run --cidr 128.2.0.0/16  --resolvers 1.1.1.1 --qps 500 # CMU
	python3 ../rdns.py run --cidr 128.237.0.0/22  --resolvers 1.1.1.1 --qps 500 # CMU
	python3 ../rdns.py run --cidr 46.101.0.0/16 --resolvers 8.8.8.8 --qps 500 # digital ocean
	python3 ../rdns.py run --cidr 140.247.0.0/16  --resolvers 1.1.1.1 --qps 500 # harvard
	python3 ../rdns.py run --cidr 128.8.0.0/16 --resolvers 1.1.1.1 --qps 500 # UMD
	python3 ../rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 # UMD
	python3 ../rdns.py run --cidr 129.2.0.0/16  --resolvers 1.1.1.1 --qps 500 # UMD
	python3 ../rdns.py run --cidr 147.222.0.0/16  --resolvers 1.1.1.1 --qps 500 # Gonzaga University


# Resolvers
```json
HE: 74.82.42.42

D-root: 199.7.91.13 and 199.7.83.42 // elapsedTime = 7.6032842 , computed QPS= 33.669660802630524
Verisign: 64.6.64.6  // elapsedTime = 12.1183899 , computed QPS= 21.12491858344977
Level 3: 4.2.2.1  // elapsedTime = 12.2512783 , computed QPS= 20.89577868784517
Neustar UltraDNS: 156.154.70.1 156.154.71.1 // elapsedTime = 9.3187372 , computed QPS= 27.471533374715193
Norton ConnectSafe: 199.85.126.10 //elapsedTime = 11.7808889 , computed QPS= 21.730109007309284
cloudflare: 1.1.1.1 and 1.0.0.1
google: 8.8.8.8 and 8.8.4.4
OpenDNS: 208.67.222.222 and 208.67.220.220  // elapsedTime = 10.6234269 , computed QPS= 24.097685465318165

// To slow
Quad 9: 9.9.9.9 and 149.112.112.112
Cleanbrowsing: 185.228.168.9 and 185.228.169.9. // 10 qps
Comodo: 8.26.56.26 and 8.20.247.20
DNS.Watch: 84.200.69.80

// works but times out after two runs of a /16 at 300 qps
Quad9: 9.9.9.9
SafeDNS: 195.46.39.39
Cleaning Browsing: 185.228.168.9

// Comodo: 8.26.56.26. lasted like 4-5 runs
```

# Num /24s in CIDR 
/0	16,777,216
/1	8,388,608
/2	4,194,304
/3	2,097,152
/4	1,048,576
/5	524,288
/6	262,144
/7	131,072
/8	65,536
/9	32,768
/10	16,384
/11	8,192
/12	4,096
/13	2,048
/14	1,024
/15	512
/16	256
/17	128
/18	64
/19	32
/20	16
/21	8
/22	4
/23	2
/24 1

# ChatGPT 
List another 100 public resolves in that format. The resolvers must be new and not previously given. If the ip Varies, specify those ips as its own result. Return in this format "ServiceName, ip1, ip2". Remove any spaces in the service name.

# DNS query
echo "Comodo Secure DNS,8.26.56.26" ; dig @8.26.56.26 -x 128.8.0.0 +noall +answer +stats +timeout=1 | awk '/Query time:/ {print $4} /^[^;]/ {print $NF}' ; echo ""

# Sort by third column 
sort -nk3 -t'-' test.txt