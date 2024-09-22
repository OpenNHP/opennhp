#/bin/bash

#	clear iptables firewall before starting up nhp-door
#
iptables -F
iptables -X
sleep 1

CURRENT_DIR=`cd \`dirname $0\`; pwd`

#
#	set iptables firewall DROP
#
echo "Setting ipset"
ipset -exist create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 130
ipset -exist create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 132
ipset -exist create tempset hash:net,port counters maxelem 1000000 timeout 10

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