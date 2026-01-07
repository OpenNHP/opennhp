#!/bin/bash
CURRENT_DIR=$(cd "$(dirname "$0")" && pwd)
if [ "$1" = "-f" ];then
    echo "Flushing existing iptables rules..."
    iptables -F
    iptables -X
fi
### ipset (IPv4) ###
echo "Setting up IPv4 ipset"
echo ""
ipset -exist create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 120
ipset -exist create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 121
ipset -exist create tempset hash:net,port counters maxelem 1000000 timeout 5
echo ""
echo "Setting IPv4 ipset OK ..."

### ipset (IPv6) ###
IP6TABLES=$(which ip6tables 2>/dev/null)
IPSET6_OK=0
if [ -n "$IP6TABLES" ]; then
    echo "Setting up IPv6 ipset"
    ipset -exist create defaultset_v6 hash:ip,port,ip family inet6 counters maxelem 1000000 timeout 120 2>/dev/null || true
    ipset -exist create defaultset_down_v6 hash:ip,port,ip family inet6 counters maxelem 1000000 timeout 121 2>/dev/null || true
    ipset -exist create tempset_v6 hash:net,port family inet6 counters maxelem 1000000 timeout 5 2>/dev/null || true

    # Verify IPv6 ipset creation
    IPSET6_OK=1
    ipset list defaultset_v6 > /dev/null 2>&1 || IPSET6_OK=0
    if [ $IPSET6_OK -eq 1 ]; then
        echo "Setting IPv6 ipset OK ..."
    fi
fi

### NHP_BLOCK chain ###
echo "Setting up NHP_BLOCK chain ..."
echo ""
iptables -N NHP_BLOCK
iptables -C NHP_BLOCK -j LOG --log-prefix "[NHP-BLOCK] " --log-level 6 --log-ip-options > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_BLOCK -j LOG --log-prefix "[NHP-BLOCK] " --log-level 6 --log-ip-options
fi
iptables -C NHP_BLOCK -j DROP > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_BLOCK -j DROP
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
# iptables -C INPUT -p tcp --dport 22  -j ACCEPT > /dev/null 2>&1
# if [ $? -ne 0 ]; then
#     iptables -I INPUT -p tcp --dport 22  -j ACCEPT
# fi

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

### chain policy (IPv4) ###
iptables -P INPUT DROP
iptables -P OUTPUT ACCEPT
iptables -P FORWARD DROP

### IPv6 firewall rules ###
if [ -n "$IP6TABLES" ] && [ $IPSET6_OK -eq 1 ]; then
    echo "Setting up IPv6 NHP_BLOCK chain ..."
    ip6tables -N NHP_BLOCK 2>/dev/null || true
    ip6tables -C NHP_BLOCK -j LOG --log-prefix "[NHP-BLOCK6] " --log-level 6 --log-ip-options > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A NHP_BLOCK -j LOG --log-prefix "[NHP-BLOCK6] " --log-level 6 --log-ip-options 2>/dev/null || true
    fi
    ip6tables -C NHP_BLOCK -j DROP > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A NHP_BLOCK -j DROP 2>/dev/null || true
    fi

    echo "Setting up IPv6 INPUT chain ..."
    # tempset_v6 -> defaultset_v6
    ip6tables -C INPUT -m set --match-set tempset_v6 src,dst -j SET --add-set defaultset_v6 src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m set --match-set tempset_v6 src,dst -j SET --add-set defaultset_v6 src,dst,dst 2>/dev/null || true
    fi

    # defaultset_v6 -> defaultset_down_v6
    ip6tables -C INPUT -m set --match-set defaultset_v6 src,dst,dst -j SET --add-set defaultset_down_v6 src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m set --match-set defaultset_v6 src,dst,dst -j SET --add-set defaultset_down_v6 src,dst,dst 2>/dev/null || true
    fi

    # defaultset_v6 accept with logging
    ip6tables -C INPUT -m set --match-set defaultset_v6 src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT6] " --log-level 6 --log-ip-options > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m set --match-set defaultset_v6 src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT6] " --log-level 6 --log-ip-options 2>/dev/null || true
    fi
    ip6tables -C INPUT -m set --match-set defaultset_v6 src,dst,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m set --match-set defaultset_v6 src,dst,dst -j ACCEPT 2>/dev/null || true
    fi

    # tempset_v6 accept
    ip6tables -C INPUT -m set --match-set tempset_v6 src,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m set --match-set tempset_v6 src,dst -j ACCEPT 2>/dev/null || true
    fi

    # loopback interface
    ip6tables -C INPUT -i lo -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -I INPUT -i lo -j ACCEPT
    fi

    # established connections
    ip6tables -C INPUT -m state --state ESTABLISHED -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -m state --state ESTABLISHED -j ACCEPT
    fi

    # rest of INPUT
    ip6tables -C INPUT -j NHP_BLOCK > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A INPUT -j NHP_BLOCK 2>/dev/null || true
    fi

    echo "Setting up IPv6 FORWARD chain ..."
    # defaultset_v6 -> defaultset_down_v6
    ip6tables -C FORWARD -m set --match-set defaultset_v6 src,dst,dst -j SET --add-set defaultset_down_v6 src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A FORWARD -m set --match-set defaultset_v6 src,dst,dst -j SET --add-set defaultset_down_v6 src,dst,dst 2>/dev/null || true
    fi

    # defaultset_v6 forward with logging
    ip6tables -C FORWARD -m set --match-set defaultset_v6 src,dst,dst -j LOG --log-prefix "[NHP-FORWARD6] " --log-level 6 --log-ip-options > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A FORWARD -m set --match-set defaultset_v6 src,dst,dst -j LOG --log-prefix "[NHP-FORWARD6] " --log-level 6 --log-ip-options 2>/dev/null || true
    fi
    ip6tables -C FORWARD -m set --match-set defaultset_v6 src,dst,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A FORWARD -m set --match-set defaultset_v6 src,dst,dst -j ACCEPT 2>/dev/null || true
    fi

    # established connections
    ip6tables -C FORWARD -m state --state ESTABLISHED -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A FORWARD -m state --state ESTABLISHED -j ACCEPT
    fi

    # rest of FORWARD
    ip6tables -C FORWARD -j NHP_BLOCK > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        ip6tables -A FORWARD -j NHP_BLOCK 2>/dev/null || true
    fi

    ### IPv6 chain policy ###
    ip6tables -P INPUT DROP
    ip6tables -P OUTPUT ACCEPT
    ip6tables -P FORWARD DROP

    echo "Setting IPv6 iptables OK ..."
fi

### iptables kernel logging ###
if [ -d /etc/rsyslog.d ] && [ ! -f /etc/rsyslog.d/10-nhplog.conf ]; then
    echo "Setting up rsyslog ..."
    mkdir -p logs
    chmod -R 777 logs/
    echo ":msg,contains,\"[NHP-ACCEPT]\" -$CURRENT_DIR/logs/nhp_accept.log

& stop
:msg,contains,\"[NHP-FORWARD]\" -$CURRENT_DIR/logs/nhp_forward.log

& stop
:msg,contains,\"[NHP-BLOCK]\" -$CURRENT_DIR/logs/nhp_block.log

& stop" > /etc/rsyslog.d/10-nhplog.conf
    systemctl restart rsyslog
fi

echo "Setting iptables default OK ..."
echo ""
### EOF ###