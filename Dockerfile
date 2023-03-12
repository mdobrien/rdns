FROM ubuntu:bionic
# LABEL maintainer="mike <obrienmd4@protonmail.com>"

ENV DEBIAN_FRONTEND noninteractive
ENV TERM xterm

RUN apt-get update -y \ 
 && apt-get install -qq -y iputils-ping dnsutils net-tools nano gcc git python3.8 python3.8-dev python3-distutils python3-setuptools \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

RUN rm -rf /usr/bin/python3 \
 && ln -s /usr/bin/python3.8 /usr/bin/python3

COPY get-pip.py /tmp/get-pip.py
RUN python3 /tmp/get-pip.py \ 
 && rm /tmp/get-pip.py

ARG PIP_INDEX_URL
ENV PIP_INDEX_URL=$PIP_INDEX_URL
ARG PIP_TRUSTED_HOST
ENV PIP_TRUSTED_HOST=$PIP_TRUSTED_HOST

RUN pip3 --no-cache install --upgrade pip setuptools wheel scapy

