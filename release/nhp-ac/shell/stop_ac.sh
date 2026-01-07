#!/bin/bash

#	clear iptables firewall before stopping up fwknopd
#
# Clear IPv4 iptables
iptables -P INPUT ACCEPT -w
iptables -P OUTPUT ACCEPT -w
iptables -P FORWARD ACCEPT -w
iptables -F -w
iptables -X -w
echo "Clear IPv4 iptables OK..."

# Clear IPv6 iptables
ip6tables -P INPUT ACCEPT -w 2>/dev/null || true
ip6tables -P OUTPUT ACCEPT -w 2>/dev/null || true
ip6tables -P FORWARD ACCEPT -w 2>/dev/null || true
ip6tables -F -w 2>/dev/null || true
ip6tables -X -w 2>/dev/null || true
echo "Clear IPv6 iptables OK..."
echo ""
sleep 1

PROC_NAME=doord
killall -9 $PROC_NAME
sleep 1

# Destroy IPv4 ipsets
ipset destroy defaultset 2>/dev/null || true
ipset destroy defaultset_down 2>/dev/null || true
ipset destroy tempset 2>/dev/null || true

# Destroy IPv6 ipsets
ipset destroy defaultset_v6 2>/dev/null || true
ipset destroy defaultset_down_v6 2>/dev/null || true
ipset destroy tempset_v6 2>/dev/null || true
echo "Clear ipsets OK..."

ProcNumber=`ps -aux | grep $PROC_NAME | grep -v grep | wc -l`
if [ $ProcNumber -le 0 ]
then
	echo "Stop NHP-Door successful..."
	echo ""
else
	echo "Stop NHP-Door fail..."
	echo ""
fi

exit