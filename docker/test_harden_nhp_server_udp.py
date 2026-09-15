"""Run with python3 docker/test_harden_nhp_server_udp.py; no firewall changes."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class FirewallRulesTest(unittest.TestCase):
    def test_rules_and_input_validation(self):
        script = Path(__file__).with_name("harden_nhp_server_udp.sh")
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            log = root / "calls"
            for tool in ("iptables", "ip6tables", "iptables-restore", "ip6tables-restore", "sysctl"):
                stub = root / tool
                stub.write_text('#!/bin/sh\nname=${0##*/}\nprintf "%s %s\\n" "$name" "$*" >> "$CALL_LOG"\ncase "$name" in\n *restore)\n  sed "s/^/$name /" >> "$CALL_LOG"\n  if [ "$name" = "ip6tables-restore" ] && [ "$1" = "--wait" ] && [ "${FAIL_IPV6:-0}" = 1 ]; then exit 1; fi;;\n sysctl) echo "${RMEM_MAX:-134217728}";;\n *) echo probe-output; exit "${HOOK_EXISTS:-0}";;\nesac\n')
                stub.chmod(0o700)
            env = dict(os.environ, PATH=f"{root}:{os.environ['PATH']}", CALL_LOG=str(log),
                       NHP_TRUSTED_PEERS="192.0.2.10 2001:db8::10")
            subprocess.run(["bash", str(script)], env=env, check=True, capture_output=True)
            calls = log.read_text().splitlines()
            for tool, peer in (("iptables", "192.0.2.10/32"), ("ip6tables", "2001:db8::10/128")):
                rules = [line for line in calls if line.startswith(tool + "-restore -A")]
                self.assertIn("-i lo -j RETURN", rules[0])
                self.assertIn(f"-s {peer} -j RETURN", rules[1])
                self.assertIn("-m limit", rules[2])
                self.assertTrue(rules[3].endswith("-j DROP"))
                self.assertFalse(any("hashlimit" in line for line in rules))
            self.assertFalse(any("sysctl -w" in line for line in calls))
            self.assertFalse(any("-I INPUT" in line for line in calls))
            log.write_text("")
            subprocess.run(["bash", str(script)], env=dict(env, HOOK_EXISTS="1", RMEM_MAX="212992"), check=True, capture_output=True)
            calls = log.read_text()
            self.assertIn("-I INPUT 1 -p udp --dport 62206 -j NHP_KNOCK_GUARD", calls)
            self.assertIn("sysctl -w net.core.rmem_max=8388608", calls)
            self.assertNotIn("probe-output", calls)
            result = subprocess.run(["bash", str(script)], env=dict(env, FAIL_IPV6="1"), capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("IPv4 applied, IPv6 FAILED", result.stderr)
            for override in ({"NHP_TRUSTED_PEERS": ""}, {"NHP_TRUSTED_PEERS": "0.0.0.0/0"},
                             {"NHP_TRUSTED_PEERS": "192.0.2.10/24"},
                             {"NHP_TRUSTED_PEERS": "5"},
                             {"NHP_KNOCK_GLOBAL_RATE_BURST": "900000000"},
                             {"NHP_TRUSTED_PEERS": "192.0.2.10 invalid"},
                             {"NHP_UDP_RECV_BUFFER_BYTES": "4294967296"},
                             {"NHP_KNOCK_PORT": "99999999999999999999999"}):
                log.write_text("")
                result = subprocess.run(["bash", str(script)], env=dict(env, **override), capture_output=True)
                self.assertNotEqual(result.returncode, 0, override)
                self.assertEqual(log.read_text(), "", override)


if __name__ == "__main__":
    unittest.main()
