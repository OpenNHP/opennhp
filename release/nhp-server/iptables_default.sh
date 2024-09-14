### INPUT chain ###
echo "Setting up INPUT chain ..."
echo ""

### Add a rule at the end of INPUT: reject ICMP for IP packets entering the udp 62206 port and less than 160 bytes in size, to prevent the port from being scanned.
iptables -C INPUT -p udp --dport 62206 -m length --length 0:160 -j REJECT --reject-with icmp-port-unreachable > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -p udp --dport 62206 -m length --length 0:160 -j REJECT --reject-with icmp-port-unreachable
fi

echo "Setting iptables default OK ..."
echo ""
