#!/bin/bash

#	clear iptables firewall before starting up nhp-door
#
# Clear IPv4 iptables
iptables -F
iptables -X

# Clear IPv6 iptables (if available)
ip6tables -F 2>/dev/null || true
ip6tables -X 2>/dev/null || true
sleep 1

CURRENT_DIR=`cd \`dirname $0\`; pwd`

#
#	set iptables firewall DROP
#
echo "Setting IPv4 ipset"
ipset -exist create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 130
ipset -exist create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 132
ipset -exist create tempset hash:net,port counters maxelem 1000000 timeout 10

# Set up IPv6 ipsets (if ip6tables available)
IP6TABLES=$(which ip6tables 2>/dev/null)
if [ -n "$IP6TABLES" ]; then
    echo "Setting IPv6 ipset"
    ipset -exist create defaultset_v6 hash:ip,port,ip family inet6 counters maxelem 1000000 timeout 130 2>/dev/null || true
    ipset -exist create defaultset_down_v6 hash:ip,port,ip family inet6 counters maxelem 1000000 timeout 132 2>/dev/null || true
    ipset -exist create tempset_v6 hash:net,port family inet6 counters maxelem 1000000 timeout 10 2>/dev/null || true
fi

source $CURRENT_DIR/iptables_default.sh

sleep 1

nohup /project/nhp-door/bin/doord run >> /project/nhp-door/shell/logs/door_start.log 2>&1 &

echo "Ready to run NHP-Door Server ..."
echo ""
sleep 1

#
#	check if NHP-Door is runing
#

PROC_NAME=doord
ProcNumber=`ps -aux | grep $PROC_NAME | grep -v color | wc -l`
if [ $ProcNumber -ge 1 ];then
   echo "Staring NHP-Door successful..."
else
   echo "Staring NHP-Door fail..."
   echo ""
fi

exit