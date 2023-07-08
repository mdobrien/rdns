#!/bin/bash

alias ipt='curl ifconfig.me; echo'

IP1=$(curl ifconfig.me)
echo "Real IP: $IP1"


openvpn --config /vpn_config/openvpn.ovpn &
cp /resolv.conf /etc/resolv.conf

IP2=$(curl ifconfig.me)
echo "VPN IP: $IP2"