### INPUT chain ###
echo "Setting up INPUT chain ..."
echo ""

### 在INPUT的最后加一条规则：对进入udp 62206端口并小于160字节的IP包进行icmp拒绝处理，防止端口被扫描到
iptables -C INPUT -p udp --dport 62206 -m length --length 0:160 -j REJECT --reject-with icmp-port-unreachable > /dev/null 2>&1
if [ $? -ne 0 ]; then
    iptables -A INPUT -p udp --dport 62206 -m length --length 0:160 -j REJECT --reject-with icmp-port-unreachable
fi

echo "Setting iptables default OK ..."
echo ""
