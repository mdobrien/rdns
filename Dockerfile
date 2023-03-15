FROM mdobrien/python:latest
# LABEL maintainer="mike <obrienmd4@protonmail.com>"

ENV DEBIAN_FRONTEND noninteractive
ENV TERM xterm

RUN apt-get update -y \ 
 && apt-get install -qq -y tcpdump iputils-ping dnsutils net-tools nano gcc git \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*


# RUN pip3 --no-cache install scapy aiodns \
RUN pip3 --no-cache install scapy \
    && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# Copy in source
COPY rdns.py /root/rdns.py

WORKDIR /root/
# ENTRYPOINT ['python3', '/root/rdns.py']

