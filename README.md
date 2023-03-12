# rdns



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