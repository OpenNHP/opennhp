"""Check the download boundary without host, package or firewall changes."""
import hashlib
from pathlib import Path
import shlex
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class PreparationTest(unittest.TestCase):
    def test_download_requires_actual_qualified_kernel_and_owner(self):
        script = (ROOT / 'scripts/prepare-kernel-containment.sh').read_text()
        cases = [
            (False, '6.1.189-236.400.amzn2023.x86_64', 'kernel-6.1.189-236.400.amzn2023.x86_64', 0, 0, True),
            (False, '6.12.113-138.209.amzn2023.x86_64', 'kernel6.12-6.12.113-138.209.amzn2023.x86_64', 0, 0, True),
            (False, '6.12.115-140.210.amzn2023.x86_64', 'kernel6.12-6.12.115-140.210.amzn2023.x86_64', 0, 0, True),
            (False, '6.12.115-140.210.amzn2023.x86_64', 'kernel6.12-6.12.113-138.209.amzn2023.x86_64', 0, 0, False),
            (False, '6.12.115-140.210.amzn2023.x86_64', 'kernel6.12-6.12.115-140.210.amzn2023.x86_64\nother', 0, 0, False),
            (False, '6.12.115-140.210.amzn2023.x86_64', 'kernel6.12-6.12.115-140.210.amzn2023.x86_64', 1, 0, False),
            (False, '6.12.116-141.211.amzn2023.x86_64', 'kernel6.12-6.12.116-141.211.amzn2023.x86_64', 0, 0, False),
            (False, '6.1.188-233.393.amzn2023.x86_64', 'kernel-old', 0, 0, False),
            (False, '6.1.189-236.400.amzn2023.x86_64', 'kernel-wrong', 0, 0, False),
            (False, '6.1.189-236.400.amzn2023.x86_64', 'kernel-6.1.189-236.400.amzn2023.x86_64\nother', 0, 0, False),
            (False, '6.1.189-236.400.amzn2023.x86_64', 'kernel-6.1.189-236.400.amzn2023.x86_64', 1, 0, False),
            (False, '6.1.189-236.400.amzn2023.x86_64', 'kernel-6.1.189-236.400.amzn2023.x86_64', 0, 1, False),
            (True, 'old-kernel', 'old-owner', 0, 0, True),
        ]
        for tools, kernel, owner, rpm_status, dnf_status, allowed in cases:
            with self.subTest(kernel=kernel, owner=owner, tools=tools), tempfile.TemporaryDirectory() as tmp:
                trace = Path(tmp) / 'trace'
                stubs = f"""
command() {{ if [[ "$*" == '-v iptables' ]]; then return {0 if tools else 1}; fi; builtin command "$@"; }}
sysctl() {{ [[ "$*" == '-w net.mptcp.enabled=0' ]] || return 89; }}
uname() {{ [[ "$*" == '-r' ]] || return 89; printf '%s\\n' {shlex.quote(kernel)}; }}
rpm() {{
  [[ "$*" == {shlex.quote('-qf /boot/vmlinuz-' + kernel + ' --qf %{NAME}-%{VERSION}-%{RELEASE}.%{ARCH}')} ]] || return 89
  printf '%s\\n' {shlex.quote(owner)}; return {rpm_status}
}}
dnf() {{ [[ "$*" == 'install -y iptables-nft' ]] || return 89; touch {shlex.quote(str(trace))}; return {dnf_status}; }}
"""
                result = subprocess.run(['bash', '-c', stubs + script], capture_output=True)
                self.assertEqual(result.returncode == 0, allowed, result.stderr)
                self.assertEqual(trace.exists(), not tools and (allowed or dnf_status == 1))

    def test_all_six_callers_prepare_before_original_installer(self):
        installer = ROOT / 'scripts/mitigate-cve-2026-98432.sh'
        self.assertEqual(hashlib.sha256(installer.read_bytes()).hexdigest(),
                         '60e172fd9ed5f66e035124d31732557e12692b1333ce49e1fdb478057245583c')
        ec2 = (ROOT / 'terraform/demo/ec2.tf').read_text()
        self.assertEqual(ec2.count('kernel_preparation_base64 = filebase64('), 3)
        for role in ('server', 'ac', 'relay'):
            script = (ROOT / f'terraform/demo/userdata/{role}.sh').read_text()
            self.assertLess(script.index('bash /run/layerv-kernel-containment-prepare.sh'),
                            script.index('bash /run/layerv-kernel-mitigation-bootstrap.sh'))
            self.assertNotIn('dnf install', script.split('bash /run/layerv-kernel-mitigation-bootstrap.sh')[0])
        workflow = (ROOT / '.github/workflows/deploy-demo-v2.yml').read_text()
        steps = workflow.split('- name: Apply and verify CVE-2026-98432 containment')[1:]
        self.assertEqual(len(steps), 3)
        for step in steps:
            step = step.split('\n      - name:', 1)[0]
            self.assertNotIn('dnf install', step)
            self.assertLess(step.index("'sudo bash -s' < repo/scripts/prepare-kernel-containment.sh"),
                            step.index("'set -eu"))


if __name__ == '__main__':
    unittest.main()
