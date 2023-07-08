# LABEL maintainer="mike <obrienmd4@protonmail.com>"
FROM mdobrien/python:latest

ENV DEBIAN_FRONTEND noninteractive
ENV TERM xterm
ENV GOVERSION "1.20.3"

RUN apt-get update -y \
 && apt-get install -qq -y curl openvpn tcpdump iputils-ping dnsutils net-tools nano gcc git gdb \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/* 

RUN pip3 --no-cache install scapy \
    && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

RUN mkdir /data

# Install golang
RUN curl -L https://golang.org/dl/go$GOVERSION.linux-amd64.tar.gz -o /tmp/go$GOVERSION.linux-amd64.tar.gz \
 && mkdir /go \
 && tar -C / -xvf /tmp/go$GOVERSION.linux-amd64.tar.gz \
 && rm /tmp/*

ENV PATH="$PATH:/go/bin"


RUN mkdir -p /rdns/src/
RUN mkdir -p /rdns/bin
RUN mkdir -p /rdns/scripts
# RUN mkdir -p /vpn_configs/ch

COPY rdns.py /rdns/src/
COPY scripts/* /rdns/scripts/
RUN chmod +x /rdns/scripts/*
COPY godns/* /rdns/src/
# COPY vpn_configs/ch-1/* /vpn_configs/

WORKDIR /rdns/src

RUN go get github.com/Workiva/go-datastructures && \
 go get github.com/miekg/dns && \
 go get github.com/cornelk/hashmap && \
 go get github.com/shogo82148/go-shuffle && \
 go get golang.org/x/net && \
 go get golang.org/x/sys


RUN echo "alias rdns.py='python3 /rdns/src/rdns.py'" >> /root/.bashrc
RUN echo "alias ec='rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500'"  >> /root/.bashrc
RUN echo "alias eg='rdns.py run --cidr 128.8.0.0/24 --resolvers 8.8.8.8 --qps 500'" >> /root/.bashrc
RUN echo "alias rb='cd /rdns/src/ && go get && (rm /rdns/bin/dns || /bin/true) && go build -gcflags "-N "  -o /rdns/bin/dns /rdns/src/dns.go'" >> /root/.bashrc
RUN echo "alias tc='grep Finished /root/debug.log  | cat -n | tail -n 1'" >> /root/.bashrc
RUN echo "alias dstats='/rdns/scripts/stats.sh'" >> /root/.bashrc
RUN echo "alias ipt='curl ifconfig.me; echo'" >> /root/.bashrc
RUN echo "alias vpn='openvpn --config /vpn_config/openvpn.ovpn &'" >> /root/.bashrc
RUN echo "alias setup='/rdns/scripts/setup.sh'" >> /root/.bashrc
RUN echo "alias d1='python3 rdns.py run  --cidr 128.0.0.0/10 --resolvers 114.114.114.114 114.114.115.115 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 9.9.9.9 --qps 500'" >> /root/.bashrc
RUN echo "alias d2='python3 rdns.py run  --cidr 128.64.0.0/10 --resolvers 114.114.114.114 114.114.115.115 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 9.9.9.9 --qps 500'" >> /root/.bashrc
RUN echo "alias d3='python3 rdns.py run  --cidr 128.128.0.0/10 --resolvers 114.114.114.114 114.114.115.115 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 9.9.9.9 --qps 500'" >> /root/.bashrc
RUN echo "alias d4='python3 rdns.py run  --cidr 128.192.0.0/10 --resolvers 114.114.114.114 114.114.115.115 64.6.64.6 156.154.70.1 199.85.126.10 1.1.1.1 1.0.0.1 8.8.8.8 8.8.4.4 208.67.222.222 208.67.220.220 9.9.9.9 --qps 500'" >> /root/.bashrc

# RUN echo "nameserver 1.1.1.1" > /etc/resolv.conf

RUN go build -o /rdns/bin/dns /rdns/src/dns.go


# ENTRYPOINT ['python3', '/root/rdns.py']





#####################
#####################
#####################
#####################

# # LABEL maintainer="mike <obrienmd4@protonmail.com>"
# FROM qmcgaw/gluetun

# ENV DEBIAN_FRONTEND noninteractive
# ENV TERM xterm
# ENV GOVERSION "1.20.3"

# # dnsutils
# RUN apk update  \
#  && apt install go curl tcpdump iputils-ping net-tools nano gcc git gdb \
#  && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/* 

# RUN mkdir /data

# # Install golang
# # RUN curl -L https://golang.org/dl/go$GOVERSION.linux-amd64.tar.gz -o /tmp/go$GOVERSION.linux-amd64.tar.gz \
# #  && mkdir /go \
# #  && tar -C / -xvf /tmp/go$GOVERSION.linux-amd64.tar.gz \
# #  && rm /tmp/*

# # ENV PATH="/go/bin:$PATH"
# # ENV PATH="$PATH:/go/bin"


# RUN mkdir -p /rdns/src/
# RUN mkdir -p /rdns/bin
# RUN mkdir -p /rdns/scripts

# COPY rdns.py /rdns/src/
# COPY scripts/stats.sh /rdns/scripts
# COPY godns/* /rdns/src/

# WORKDIR /rdns/src

# RUN go get github.com/Workiva/go-datastructures && \
#  go get github.com/miekg/dns && \
#  go get github.com/cornelk/hashmap && \
#  go get github.com/shogo82148/go-shuffle && \
#  go get golang.org/x/net && \
#  go get golang.org/x/sys


# RUN echo "alias rdns.py='python3 /rdns/src/rdns.py'" >> /root/.bashrc
# RUN echo "alias ec='rdns.py run --cidr 128.8.0.0/24 --resolvers 1.1.1.1 --qps 500'"  >> /root/.bashrc
# RUN echo "alias eg='rdns.py run --cidr 128.8.0.0/24 --resolvers 8.8.8.8 --qps 500'" >> /root/.bashrc
# RUN echo "alias rb='cd /rdns/src/ && go get && (rm /rdns/bin/dns || /bin/true) && go build -gcflags "-N -l"  -o /rdns/bin/dns /rdns/src/dns.go'" >> /root/.bashrc
# RUN echo "alias tc='grep Finished /root/debug.log  | cat -n | tail -n 1'" >> /root/.bashrc
# RUN echo "alias dstats='/rdns/scripts/stats.sh'" >> /root/.bashrc

# RUN apt install bash curl

# RUN go build -o /rdns/bin/dns /rdns/src/dns.go

# # ENV VPN_SERVICE_PROVIDER=cyberghost 
# # ENV OPENVPN_USER=VLwhdm7R9j 
# # ENV OPENVPN_PASSWORD=MzBhCq2Tjb 
# # ENV SERVER_COUNTRIES="United States" 
# # COPY openvpn /gluetun 

# # COPY get-pip.py /tmp/get-pip.py
# # RUN python3 /tmp/get-pip.py \ 
# #  && rm /tmp/get-pip.py
# RUN apt install py3-pip

# RUN pip3 --no-cache install scapy \
#     && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*



# # ENTRYPOINT ['python3', '/root/rdns.py']
# # ENTRYPOINT ['/bin/bash']
# ENTRYPOINT ["/gluetun-entrypoint"]

