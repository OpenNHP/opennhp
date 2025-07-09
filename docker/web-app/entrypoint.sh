#!/bin/bash
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -I INPUT -p tcp --dport 8080 -s 177.7.0.10 -j ACCEPT

/app