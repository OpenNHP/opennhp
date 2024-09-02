#/bin/bash

#	clear iptables firewall before stopping up fwknopd
#
iptables -P INPUT ACCEPT -w
iptables -P OUTPUT ACCEPT -w
iptables -P FORWARD ACCEPT -w
iptables -F -w
iptables -X -w
ip6tables -F -w
ip6tables -X -w
echo "Clear iptables OK..."
echo ""
sleep 1

PROC_NAME=doord
killall -9 $PROC_NAME
sleep 1

ipset destroy defaultset
ipset destroy defaultset_down

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