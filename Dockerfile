FROM mdobrien/python:latest
# LABEL maintainer="mike <obrienmd4@protonmail.com>"

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

# RUN pip3 --no-cache install scapy aiodns \
RUN pip3 --no-cache install scapy \
    && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# Copy in source
COPY rdns.py /root/rdns.py

WORKDIR /root/
# ENTRYPOINT ['python3', '/root/rdns.py']

