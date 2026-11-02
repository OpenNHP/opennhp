#!/usr/bin/env python3
"""Safe packet test on Linux: two temporary namespaces, no exploit traffic."""
import os
import pathlib
import subprocess
import sys

def main():
    script = pathlib.Path(sys.argv[1]).resolve()
    prefix = 'lv-cve-' + str(os.getpid())
    server, client = prefix + '-s', prefix + '-c'

    def run(*args):
        return subprocess.check_output(args, text=True, stderr=subprocess.STDOUT)

    def inside(ns, *args):
        return run('ip', 'netns', 'exec', ns, *args)

    def connect(options=False, ns=client):
        code = "import socket; s=socket.socket(); s.settimeout(1); "
        if options:
            code += "s.setsockopt(socket.IPPROTO_IP,socket.IP_OPTIONS,bytes([1,1,1,0])); "
        code += "s.connect(('198.18.0.1',19843)); assert s.recv(2)==b'ok'"
        return subprocess.run(['ip', 'netns', 'exec', ns, 'python3', '-c', code],
                              capture_output=True, text=True)

    listener = None
    try:
        for ns in (server, client):
            run('ip', 'netns', 'add', ns)
            inside(ns, 'ip', 'link', 'set', 'lo', 'up')
        run('ip', 'link', 'add', 'lvcve0', 'type', 'veth', 'peer', 'name', 'lvcve1')
        for ns, dev, address in ((server, 'lvcve0', '198.18.0.1/30'),
                                 (client, 'lvcve1', '198.18.0.2/30')):
            run('ip', 'link', 'set', dev, 'netns', ns)
            inside(ns, 'ip', 'addr', 'add', address, 'dev', dev)
            inside(ns, 'ip', 'link', 'set', dev, 'up')
        code = ("import socket\n"
                "s=socket.socket(); s.bind(('198.18.0.1',19843)); s.listen()\n"
                "print('ready',flush=True)\n"
                "while True:\n c,_=s.accept(); c.sendall(b'ok'); c.close()\n")
        listener = subprocess.Popen(['ip', 'netns', 'exec', server, 'python3', '-u', '-c', code],
                                    stdout=subprocess.PIPE, text=True)
        assert listener.stdout.readline().strip() == 'ready'
        assert connect().returncode == 0
        assert connect(True).returncode == 0, 'baseline IP options must reach listener'
        assert connect(True, server).returncode == 0, 'local baseline options must work'
        for _ in range(2):
            inside(server, 'bash', str(script), '--apply')
        assert connect().returncode == 0, 'ordinary TCP must work'
        result = connect(True)
        assert result.returncode != 0 and ('TimeoutError' in result.stderr or 'socket.timeout:' in result.stderr), result.stderr
        rules = inside(server, 'iptables-save', '-c', '-t', 'raw')
        assert rules.count('--comment CVE-2026-98432') == 2, rules
        # Locally generated/loopback traffic must not bypass the mitigation.
        assert connect(True, server).returncode != 0, 'OUTPUT must drop local IP options'
        assert inside(server, 'sysctl', '-n', 'net.mptcp.enabled').strip() == '0'
        print(rules)
        print('PASS: baseline options accepted; mitigation drops options; normal TCP works; repeat apply is idempotent; MPTCP disabled')
    finally:
        if listener:
            listener.terminate()
            listener.wait(timeout=5)
        for ns in (server, client):
            subprocess.run(['ip', 'netns', 'del', ns], capture_output=True)


if __name__ == "__main__":
    main()
