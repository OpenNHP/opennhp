#!/bin/bash
CURRENT_DIR=`cd \`dirname $0\`; pwd`

### flush existing rules and set chain policy setting to DROP
if [ "$1" = "-f" ]; then
    echo "Flushing existing iptables rules..."
    echo ""
    iptables -F
    iptables -X
fi

### ipset ###
echo "Setting up ipset"
echo ""
ipset -exist create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 120
ipset -exist create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 121
ipset -exist create tempset hash:net,port counters maxelem 1000000 timeout 5
echo ""
echo "Setting ipset OK ..."

### NHP_BLOCK chain ###
echo "Setting up NHP_BLOCK chain ..."
echo ""
iptables -N NHP_BLOCK
iptables -C NHP_BLOCK -d $(hostname -I | awk '{print $1}') -j LOG --log-prefix "[NHP-BLOCK] " --log-level 6 --log-ip-options > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_BLOCK -d $(hostname -I | awk '{print $1}') -j LOG --log-prefix "[NHP-BLOCK] " --log-level 6 --log-ip-options
fi

iptables -C NHP_BLOCK -d $(hostname -I | awk '{print $1}') -j DROP > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_BLOCK -d $(hostname -I | awk '{print $1}') -j DROP
fi


### INPUT chain ###
echo "Setting up INPUT chain ..."
echo ""
# tempset -> defaultset
iptables -C INPUT -m set --match-set tempset src,dst -j SET --add-set defaultset src,dst,dst > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m set --match-set tempset src,dst -j SET --add-set defaultset src,dst,dst
fi

# defaultset -> defaultset_down
iptables -C INPUT -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst
fi

# defaultset
iptables -C INPUT -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT] " --log-level 6 --log-ip-options > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT] " --log-level 6 --log-ip-options
fi
iptables -C INPUT -m set --match-set defaultset src,dst,dst -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m set --match-set defaultset src,dst,dst -j ACCEPT
fi

# tempset
iptables -C INPUT -m set --match-set tempset src,dst -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m set --match-set tempset src,dst -j ACCEPT
fi

# loopback interface
iptables -C INPUT -i lo -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -I INPUT -i lo -j ACCEPT
fi

# ssh
iptables -C INPUT -p tcp --dport 22  -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -I INPUT -p tcp --dport 22  -j ACCEPT
fi

# established connections
iptables -C INPUT -m state --state ESTABLISHED -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -m state --state ESTABLISHED -j ACCEPT
fi

# rest of INPUT
iptables -C INPUT -j NHP_BLOCK > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -j NHP_BLOCK
fi

### OUTPUT chain ###
echo "Setting up OUTPUT chain ..."
echo ""
#iptables -A OUTPUT -m set --match-set defaultset_down dst,src,src -j SET --add-set defaultset_down dst,src,src

### FORWARD chain ###
echo "Setting up FORWARD chain ..."
echo ""

# defaultset -> defaultset_down
iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst
fi

# defaultset
iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-FORWARD] " --log-level 6 --log-ip-options > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-FORWARD] " --log-level 6 --log-ip-options
fi
iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j ACCEPT
fi

# established connections
iptables -C FORWARD -m state --state ESTABLISHED -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -m state --state ESTABLISHED -j ACCEPT
fi

# rest of FORWARD
iptables -C FORWARD -j NHP_BLOCK > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -j NHP_BLOCK
fi

### chain policy ###
iptables -P INPUT DROP
iptables -P OUTPUT ACCEPT
iptables -P FORWARD DROP

### iptables kernel logging ###
LOCAL_IP=$(ip route get 1 | awk '{print $7;exit}')
[ -z "$LOCAL_IP" ] && LOCAL_IP=$(hostname -I | awk '{print $1}')

if [ -d /etc/rsyslog.d ] && [ ! -f /etc/rsyslog.d/10-nhplog.conf ]; then
    echo "Setting up rsyslog with dynamic IP..."
    mkdir -p $CURRENT_DIR/logs
    chown $(whoami):$(id -gn) $CURRENT_DIR/logs
    chmod -R 755 $CURRENT_DIR/logs/
    setenforce 
    echo 'template(name="NHPFormat" type="string" string="%timegenerated% '"${LOCAL_IP}"' kernel: %msg%\n")

:msg,contains,"[NHP-ACCEPT]" -'"$CURRENT_DIR"'/logs/nhp_accept.log;NHPFormat
& stop
:msg,contains,"[NHP-FORWARD]" -'"$CURRENT_DIR"'/logs/nhp_forward.log;NHPFormat
& stop
:msg,contains,"[NHP-BLOCK]" -'"$CURRENT_DIR"'/logs/nhp_block.log;NHPFormat
& stop' > /etc/rsyslog.d/10-nhplog.conf

    systemctl restart rsyslog
fi

echo "Setting iptables default OK ..."
echo ""
### EOF ###
