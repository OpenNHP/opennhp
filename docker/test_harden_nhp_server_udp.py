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
            for tool in ("iptables", "ip6tables", "sysctl"):
                stub = root / tool
                stub.write_text('#!/bin/sh\nprintf "%s %s\\n" "${0##*/}" "$*" >> "$CALL_LOG"\n')
                stub.chmod(0o700)
            env = dict(os.environ, PATH=f"{root}:{os.environ['PATH']}", CALL_LOG=str(log),
                       NHP_TRUSTED_PEERS="192.0.2.10 2001:db8::10")
            subprocess.run(["bash", str(script)], env=env, check=True, capture_output=True)
            calls = log.read_text().splitlines()
            for tool, peer in (("iptables", "192.0.2.10/32"), ("ip6tables", "2001:db8::10/128")):
                rules = [line for line in calls if line.startswith(tool + " -w -A")]
                self.assertIn(f"-s {peer} -j RETURN", rules[0])
                self.assertIn("-m limit", rules[1])
                self.assertTrue(rules[2].endswith("-j DROP"))
                self.assertFalse(any("hashlimit" in line for line in rules))
            for override in ({"NHP_TRUSTED_PEERS": ""}, {"NHP_TRUSTED_PEERS": "0.0.0.0/0"},
                             {"NHP_TRUSTED_PEERS": "192.0.2.10 invalid"},
                             {"NHP_UDP_RECV_BUFFER_BYTES": "4294967296"},
                             {"NHP_KNOCK_PORT": "99999999999999999999999"}):
                log.write_text("")
                result = subprocess.run(["bash", str(script)], env=dict(env, **override), capture_output=True)
                self.assertNotEqual(result.returncode, 0, override)
                self.assertEqual(log.read_text(), "", override)


if __name__ == "__main__":
    unittest.main()
