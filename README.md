# rdns

# IDEAS
 - The goal is to create a distrubuted systems that can dynimcally scale number of works based
 IPV4 spaces, desired time to have tasking completed, queries per second
 - Create service to discover public resolvers executed recursive queries. This ips of ther resolves will then be used by the workers for executing DNS queries 
 - 

# Running
```bash
# Set project env vars and aliases
export RDNS_HOME='/path/to/rdns'
alias db='cd $RDNS_HOME'
alias biz='docker run -ti --volume $RDNS_HOME/rdns.py:/root/rdns.py rdns bash'

# Build container
docker build -t rdns .

# run container
docker run -ti --volume $RDNS_HOME/rdns.py:/root/rdns.py rdns bash


# example run
python3 rdns.py run --cidr 128.8.0.0/24 --destination 1.1.1.1 --qps 500

```


# Current stats
```
root@8a50181e61ca:~# python3 rdns.py run --cidr 128.8.0.0/24 --destination 1.1.1.1 --qps 500
#names/#queries: 66 / 256
sent DNS querys for 44.38332557678223s
writing results to /data/rdns.json
 Elapse time writing results to disk: 0.0008902549743652344s
wrote files to disk
runtime: 44.38510274887085
```


# Test 1
	python3 rdns.py run --cidr 128.8.0.0/16 --destination 1.1.1.1 --qps 500
	python3 rdns.py run --cidr 46.101.0.0/16 --destination 8.8.8.8 --qps 500