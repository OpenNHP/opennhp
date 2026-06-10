#!/bin/bash
# AC_SOURCE_IP is the NHP-AC that guards this resource; only it may reach
# port 8080. Defaults to the cluster-1 AC (177.7.0.10) so the single-cluster
# demo is unchanged. The multi-cluster demo overrides it per instance.
AC_SOURCE_IP="${AC_SOURCE_IP:-177.7.0.10}"

iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -I INPUT -p tcp --dport 8080 -s "${AC_SOURCE_IP}" -j ACCEPT

/app
