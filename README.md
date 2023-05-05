# rdns

# IDEAS
 - The goal is to create a distrubuted systems that can dynimcally scale number of works based
 IPV4 spaces, desired time to have tasking completed, queries per second
 - Create service to discover public resolvers executed recursive queries. This ips of ther resolves will then be used by the workers for executing DNS queries 
 - 

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
	rdns:test \
	/bin/bash

# example run
rdns.py run --cidr 128.8.0.0/23 --resolvers 1.1.1.1 8.8.8.8 --qps 500
rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500

```


# Current stats
```
root@8a50181e61ca:~# python3 rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500
#names/#queries: 66 / 256
sent DNS querys for 44.38332557678223s
writing results to /data/rdns.json
 Elapse time writing results to disk: 0.0008902549743652344s
wrote files to disk
runtime: 44.38510274887085
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

	# good test
	python3 ../rdns.py run --cidr 128.2.0.0/19 --resolvers 1.0.0.1 1.1.1.1 8.8.4.4 8.8.8.8 208.67.220.220 208.67.222.222 216.146.35.35 --workers 5 --qps 500

	python3 ../rdns.py run --cidr 128.8.0.0/16 --resolvers 1.0.0.1 1.1.1.1 134.195.4.2 149.112.112.112 185.228.168.9 185.228.169.9 195.46.39.39 195.46.39.40 205.171.2.65 205.171.3.65 208.67.220.220 208.67.222.222 216.146.35.35 216.146.36.36 64.6.64.6 64.6.65.6 74.82.42.42 76.223.122.150 76.76.10.0 76.76.19.19 76.76.2.0 77.88.8.1 77.88.8.8 8.20.247.20 8.26.56.26 8.8.4.4 8.8.8.8 84.200.69.80 84.200.70.40 9.9.9.9 --workers 5 --qps 500


	python3 rdns.py run --cidr 128.2.0.0/20 --resolvers 1.0.0.1 1.1.1.1 8.8.4.4 8.8.8.8 208.67.220.220 208.67.222.222 216.146.35.35 --workers 5 --qps 500

# Wil give names, nonames, and timeouts
python3 rdns.py run --cidr 128.8.0.0/24 128.8.0.0/24 --resolvers 8.8.4.4 159.89.120.99 --workers 5 --qps 500 

# /24 breakdown
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



# Resolvers
128.8.127.4 from server 1.0.0.1 in 23 ms.
128.8.127.4 from server 1.1.1.1 in 12 ms.
128.8.127.4 from server 134.195.4.2 in 74 ms.
128.8.127.4 from server 149.112.112.112 in 6 ms.
128.8.127.4 from server 185.228.168.9 in 8 ms.
128.8.127.4 from server 185.228.169.9 in 482 ms.
128.8.127.4 from server 195.46.39.39 in 8 ms.
128.8.127.4 from server 195.46.39.40 in 11 ms.
128.8.127.4 from server 205.171.2.65 in 10 ms.
128.8.127.4 from server 205.171.3.65 in 83 ms.
128.8.127.4 from server 208.67.220.220 in 15 ms.
128.8.127.4 from server 208.67.222.222 in 19 ms.
128.8.127.4 from server 216.146.35.35 in 23 ms.
128.8.127.4 from server 216.146.36.36 in 64 ms.
128.8.127.4 from server 64.6.64.6 in 81 ms.
128.8.127.4 from server 64.6.65.6 in 25 ms.
128.8.127.4 from server 74.82.42.42 in 9 ms.
128.8.127.4 from server 76.223.122.150 in 11 ms.
128.8.127.4 from server 76.76.10.0 in 21 ms.
128.8.127.4 from server 76.76.19.19 in 11 ms.
128.8.127.4 from server 76.76.2.0 in 14 ms.
128.8.127.4 from server 77.88.8.1 in 290 ms.
128.8.127.4 from server 77.88.8.8 in 344 ms.
128.8.127.4 from server 8.20.247.20 in 87 ms.
128.8.127.4 from server 8.26.56.26 in 15 ms.
128.8.127.4 from server 8.8.4.4 in 7 ms.
128.8.127.4 from server 8.8.8.8 in 16 ms.
128.8.127.4 from server 84.200.69.80 in 406 ms.
128.8.127.4 from server 84.200.70.40 in 299 ms.
128.8.127.4 from server 9.9.9.9 in 13 ms.

- does not work: dig +time=2 +short +identify @159.89.120.99 cs.umd.edu 

