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
ipset -exist create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 120 2>/dev/null || \
    ipset create defaultset hash:ip,port,ip counters maxelem 1000000 timeout 120 2>/dev/null || true
ipset -exist create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 121 2>/dev/null || \
    ipset create defaultset_down hash:ip,port,ip counters maxelem 1000000 timeout 121 2>/dev/null || true
ipset -exist create tempset hash:net,port counters maxelem 1000000 timeout 5 2>/dev/null || \
    ipset create tempset hash:net,port counters maxelem 1000000 timeout 5 2>/dev/null || true

# Verify ipset creation
IPSET_OK=1
ipset list defaultset > /dev/null 2>&1 || IPSET_OK=0
ipset list defaultset_down > /dev/null 2>&1 || IPSET_OK=0
ipset list tempset > /dev/null 2>&1 || IPSET_OK=0

if [ $IPSET_OK -eq 0 ]; then
    echo "Warning: Some ipset creation failed. Skipping ipset-related iptables rules."
    echo "This may happen on systems using nf_tables backend without ipset support."
    echo ""
fi
echo "Setting ipset OK ..."

### NHP_DENY chain ###
echo "Setting up NHP_DENY chain ..."
echo ""
iptables -N NHP_DENY
iptables -C NHP_DENY -d $(hostname -I | awk '{print $1}') -j LOG --log-prefix "[NHP-DENY] " --log-level 6 --log-ip-options > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_DENY -d $(hostname -I | awk '{print $1}') -j LOG --log-prefix "[NHP-DENY] " --log-level 6 --log-ip-options
fi

iptables -C NHP_DENY -d $(hostname -I | awk '{print $1}') -j DROP > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A NHP_DENY -d $(hostname -I | awk '{print $1}') -j DROP
fi


### INPUT chain ###
echo "Setting up INPUT chain ..."
echo ""

if [ $IPSET_OK -eq 1 ]; then
    # tempset -> defaultset
    iptables -C INPUT -m set --match-set tempset src,dst -j SET --add-set defaultset src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A INPUT -m set --match-set tempset src,dst -j SET --add-set defaultset src,dst,dst 2>/dev/null
    fi

    # defaultset -> defaultset_down
    iptables -C INPUT -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A INPUT -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst 2>/dev/null
    fi

    # defaultset
    iptables -C INPUT -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT] " --log-level 6 --log-ip-options > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A INPUT -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-ACCEPT] " --log-level 6 --log-ip-options 2>/dev/null
    fi
    iptables -C INPUT -m set --match-set defaultset src,dst,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A INPUT -m set --match-set defaultset src,dst,dst -j ACCEPT 2>/dev/null
    fi

    # tempset
    iptables -C INPUT -m set --match-set tempset src,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A INPUT -m set --match-set tempset src,dst -j ACCEPT 2>/dev/null
    fi
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
iptables -C INPUT -j NHP_DENY > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -j NHP_DENY
fi

### OUTPUT chain ###
echo "Setting up OUTPUT chain ..."
echo ""
#iptables -A OUTPUT -m set --match-set defaultset_down dst,src,src -j SET --add-set defaultset_down dst,src,src

### FORWARD chain ###
echo "Setting up FORWARD chain ..."
echo ""

if [ $IPSET_OK -eq 1 ]; then
    # defaultset -> defaultset_down
    iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j SET --add-set defaultset_down src,dst,dst 2>/dev/null
    fi

    # defaultset
    iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-FORWARD] " --log-level 6 --log-ip-options > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j LOG --log-prefix "[NHP-FORWARD] " --log-level 6 --log-ip-options 2>/dev/null
    fi
    iptables -C FORWARD -m set --match-set defaultset src,dst,dst -j ACCEPT > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        iptables -A FORWARD -m set --match-set defaultset src,dst,dst -j ACCEPT 2>/dev/null
    fi
fi

# established connections
iptables -C FORWARD -m state --state ESTABLISHED -j ACCEPT > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -m state --state ESTABLISHED -j ACCEPT
fi

# rest of FORWARD
iptables -C FORWARD -j NHP_DENY > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A FORWARD -j NHP_DENY
fi

### chain policy ###
iptables -P INPUT DROP
iptables -P OUTPUT ACCEPT
iptables -P FORWARD DROP

### iptables kernel logging ###
LOCAL_IP=$(ip route get 1 | awk '{print $7;exit}')
[ -z "$LOCAL_IP" ] && LOCAL_IP=$(hostname -I | awk '{print $1}')

if [ -d /etc/rsyslog.d ]; then
    echo "Setting up rsyslog with dynamic IP..."
    # Use /var/log/nhp for rsyslog (rsyslog runs as syslog user, cannot access /root/)
    NHP_LOG_DIR="/var/log/nhp"
    mkdir -p $NHP_LOG_DIR
    chown syslog:adm $NHP_LOG_DIR 2>/dev/null || chown root:root $NHP_LOG_DIR
    chmod 755 $NHP_LOG_DIR
    # Also create local logs directory for other uses
    mkdir -p $CURRENT_DIR/logs
    chown $(whoami):$(id -gn) $CURRENT_DIR/logs
    chmod -R 755 $CURRENT_DIR/logs/
    setenforce 0 2>/dev/null || true
    echo 'template(name="NHPFormat" type="string" string="%timegenerated:8:19% '"${LOCAL_IP}"' %syslogtag% %msg:::drop-last-lf%\n")
template(name="NHPAcceptFile" type="string" string="'"$NHP_LOG_DIR"'/nhp_accept-%$YEAR%-%$MONTH%-%$DAY%.log")
template(name="NHPForwardFile" type="string" string="'"$NHP_LOG_DIR"'/nhp_forward-%$YEAR%-%$MONTH%-%$DAY%.log")
template(name="NHPDenyFile" type="string" string="'"$NHP_LOG_DIR"'/nhp_deny-%$YEAR%-%$MONTH%-%$DAY%.log")

:msg,contains,"[NHP-ACCEPT]" ?NHPAcceptFile;NHPFormat
& stop
:msg,contains,"[NHP-FORWARD]" ?NHPForwardFile;NHPFormat
& stop
:msg,contains,"[NHP-DENY]" ?NHPDenyFile;NHPFormat
& stop' > /etc/rsyslog.d/10-nhplog.conf

    systemctl restart rsyslog
fi

### Setup daily cleanup for NHP logs (30 days retention) ###
if [ -d /etc/cron.daily ]; then
    echo "Setting up NHP logs cleanup (30 days retention)..."
    cat > /etc/cron.daily/nhp-log-cleanup << 'EOF'
#!/bin/bash
# Delete NHP log files older than 30 days
find /var/log/nhp -name "nhp_*.log" -type f -mtime +30 -delete 2>/dev/null
EOF
    chmod +x /etc/cron.daily/nhp-log-cleanup
    echo "Cron daily cleanup configured ..."
fi

echo "Setting iptables default OK ..."
echo ""
### EOF ###
