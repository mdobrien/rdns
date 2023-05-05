# LABEL maintainer="mike <obrienmd4@protonmail.com>"
FROM mdobrien/python:latest

ENV DEBIAN_FRONTEND noninteractive
ENV TERM xterm
ENV GOVERSION "1.20.3"

RUN apt-get update -y \
 && apt-get install -qq -y curl tcpdump iputils-ping dnsutils net-tools nano gcc git \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/* 

RUN mkdir /data

# Install golang
RUN curl -L https://golang.org/dl/go$GOVERSION.linux-amd64.tar.gz -o /tmp/go$GOVERSION.linux-amd64.tar.gz \
 && mkdir /go \
 && tar -C /go -xvf /tmp/go$GOVERSION.linux-amd64.tar.gz \
 && rm /tmp/*

ENV PATH="$PATH:/go/go/bin"


RUN mkdir -p /rdns/src/
RUN mkdir -p /rdns/bin

COPY godns/* /rdns/src/
WORKDIR /rdns/src

RUN go get github.com/Workiva/go-datastructures && \
 go get github.com/miekg/dns && \
 go get github.com/cornelk/hashmap && \
 go get github.com/shogo82148/go-shuffle && \
 go get golang.org/x/net && \
 go get golang.org/x/sys

RUN echo "alias rdns.py='python3 /rdns/src/rdns.py'" >> /root/.bashrc
RUN echo "alias rb='cd /rdns/src/ && go get && (rm /rdns/bin/dns || /bin/true) && go build -o /rdns/bin/dns /rdns/src/dns.go'" >> /root/.bashrc

RUN go build -o /rdns/bin/dns /rdns/src/dns.go

COPY rdns.py /rdns/src/



# RUN pip3 --no-cache install scapy aiodns \
RUN pip3 --no-cache install scapy \
    && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*



# ENTRYPOINT ['python3', '/root/rdns.py']

