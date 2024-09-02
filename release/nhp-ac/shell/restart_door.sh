#/bin/bash

#	clear iptables firewall before stopping up fwknopd
#


CURRENT_DIR=`cd \`dirname $0\`; pwd`

$CURRENT_DIR/stop_door.sh
sleep 1
$CURRENT_DIR/run_door.sh

exit