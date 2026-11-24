#!/usr/bin/env bash
# test_adopt_dns_records.sh
#
# Tests for scripts/adopt-dns-records.sh, against a stubbed Terraform and a
# stubbed Cloudflare API.
#
# The script repairs a DNS record that exists in Cloudflare but that the
# Terraform state cannot describe -- an apply that created the record and
# then died before the provider recorded its id. It writes to the state, so
# the cases that matter are the ones where it must *not* write, and the ones
# where what it writes has to be exactly the live record and nothing else.
#
# Usage:
#   tests/scripts/test_adopt_dns_records.sh [path/to/adopt-dns-records.sh]
#
# Requires: bash, jq, mktemp.

set -uo pipefail

SCRIPT_UNDER_TEST="${1:-scripts/adopt-dns-records.sh}"
if [ ! -x "$SCRIPT_UNDER_TEST" ]; then
  printf 'ERROR: %s is not an executable script.\n' "$SCRIPT_UNDER_TEST" >&2
  exit 2
fi
SCRIPT_UNDER_TEST="$(cd "$(dirname "$SCRIPT_UNDER_TEST")" && pwd)/$(basename "$SCRIPT_UNDER_TEST")"

STUB_DIR="$(mktemp -d)"
WORK_DIR="$STUB_DIR/work"
mkdir -p "$STUB_DIR/bin" "$WORK_DIR/.terraform"
trap 'rm -rf "$STUB_DIR"' EXIT

# --- the stubs ------------------------------------------------------------
#
# The Terraform stub answers the four commands the script uses and records
# every call, so a test can assert on what the script did to the state and
# not only on what it printed. The Cloudflare stub answers the zone and the
# record listing from files the test writes.

cat >"$STUB_DIR/bin/terraform" <<'STUB'
#!/usr/bin/env bash
# Stub Terraform for tests/scripts/test_adopt_dns_records.sh
MOCK_DIR="${MOCK_DIR:?}"
printf 'terraform %s\n' "$*" >>"$MOCK_DIR/calls.log"
case "$1" in
show)
  if [ "${MOCK_STATE_READ_FAIL:-0}" = "1" ] && [ -z "${3:-}" ]; then
    echo "Error: Failed to read the state" >&2
    exit 1
  fi
  if [ "$2" = "-json" ] && [ -n "${3:-}" ]; then
    cat "$3"
  else
    cat "$MOCK_DIR/state.json"
  fi
  ;;
state)
  if [ "${MOCK_STATE_READ_FAIL:-0}" = "1" ]; then
    echo "Error: Failed to read the state" >&2
    exit 1
  fi
  case "$2" in
  list) cat "$MOCK_DIR/state.list" ;;
  pull) cat "$MOCK_DIR/state.backup.json" ;;
  rm) echo "Removed $3 from the state" ;;
  push)
    path="${3:-}"
    [ "$path" = "-force" ] && path="${4:-}"
    echo "pushed $path"
    cat -- "$path" >/dev/null
    ;;
  *) exit 1 ;;
  esac
  ;;
import)
  if [ "${MOCK_IMPORT_FAIL:-0}" = "1" ]; then
    echo "Error: import failed" >&2
    exit 1
  fi
  echo "Import successful! $3 = $4"
  ;;
plan) echo "planned ${*:-}" ;;
*) echo "unexpected terraform command: $*" >&2; exit 1 ;;
esac
STUB

cat >"$STUB_DIR/bin/curl" <<'STUB'
#!/usr/bin/env bash
# Stub Cloudflare API for tests/scripts/test_adopt_dns_records.sh
MOCK_DIR="${MOCK_DIR:?}"
url="${*: -1}"
case "$url" in
*dns_records*page=2*) cat "$MOCK_DIR/dns_records.page2.json" ;;
*/dns_records*) cat "$MOCK_DIR/dns_records.json" ;;
*/zones/*) cat "$MOCK_DIR/zone.json" ;;
*) echo "unexpected url: $url" >&2; exit 1 ;;
esac
STUB

chmod +x "$STUB_DIR/bin/terraform" "$STUB_DIR/bin/curl"

export MOCK_DIR="$STUB_DIR"
export PATH="$STUB_DIR/bin:$PATH"
export CF_TOKEN="test-token"
export CF_ZONE="zone123"

# --- the fixtures ---------------------------------------------------------

# The configuration's intent: five records, one of which (demologin) is not
# live at all, and one of which (server) the state holds as a CNAME while
# the zone holds an A record -- the shape an apply that died leaves behind.
cat >"$STUB_DIR/plan.json" <<'EOF'
{
  "planned_values": {
    "root_module": {
      "resources": [
        {"address": "cloudflare_record.ac", "type": "cloudflare_record", "values": {"name": "ac", "type": "A"}},
        {"address": "cloudflare_record.server", "type": "cloudflare_record", "values": {"name": "server", "type": "A"}},
        {"address": "cloudflare_record.acdemo", "type": "cloudflare_record", "values": {"name": "acdemo", "type": "CNAME"}},
        {"address": "cloudflare_record.stats", "type": "cloudflare_record", "values": {"name": "stats", "type": "A"}},
        {"address": "cloudflare_record.demologin", "type": "cloudflare_record", "values": {"name": "demologin", "type": "CNAME"}}
      ]
    }
  },
  "variables": {"domain": {"value": "opennhp.org"}}
}
EOF

cat >"$STUB_DIR/state.json" <<'EOF'
{
  "format_version": "1.0",
  "terraform_version": "1.10.5",
  "values": {
    "root_module": {
      "resources": [
        {"address": "cloudflare_record.ac", "mode": "managed", "type": "cloudflare_record", "name": "ac", "values": {"id": "id-ac", "name": "ac", "type": "A"}},
        {"address": "cloudflare_record.acdemo", "mode": "managed", "type": "cloudflare_record", "name": "acdemo", "values": {"id": "id-acdemo", "name": "acdemo", "type": "CNAME"}},
        {"address": "cloudflare_record.server", "mode": "managed", "type": "cloudflare_record", "name": "server", "values": {"id": "id-server-cname", "name": "server", "type": "CNAME"}},
        {"address": "cloudflare_record.stats", "mode": "managed", "type": "cloudflare_record", "name": "stats", "values": {"id": "id-stats", "name": "stats", "type": "A"}}
      ]
    }
  }
}
EOF

printf 'cloudflare_record.ac\ncloudflare_record.acdemo\ncloudflare_record.server\ncloudflare_record.stats\n' >"$STUB_DIR/state.list"

# The live zone. "server" holds two records, an A and a TXT: only the A is
# the one the configuration declares, and only the A may be adopted. The
# deeper "server.sub.opennhp.org" must not be matched by "server".
cat >"$STUB_DIR/dns_records.default.json" <<'EOF'
{
  "success": true,
  "result": [
    {"id": "id-ac", "name": "ac.opennhp.org", "type": "A"},
    {"id": "id-acdemo", "name": "acdemo.opennhp.org", "type": "CNAME"},
    {"id": "id-server-a", "name": "server.opennhp.org", "type": "A"},
    {"id": "id-server-txt", "name": "server.opennhp.org", "type": "TXT"},
    {"id": "id-stats", "name": "stats.opennhp.org", "type": "A"},
    {"id": "id-deep", "name": "server.sub.opennhp.org", "type": "A"},
    {"id": "id-retired", "name": "auth-plugin.opennhp.org", "type": "A"}
  ],
  "result_info": {"total_pages": 1}
}
EOF

cat >"$STUB_DIR/zone.json" <<'EOF'
{"success": true, "result": {"id": "zone123", "name": "opennhp.org"}}
EOF

# --- the test harness -----------------------------------------------------

pass=0
fail=0

ok() {
  pass=$((pass + 1))
  printf '  ok   %s\n' "$1"
}

not_ok() {
  fail=$((fail + 1))
  printf '  FAIL %s\n' "$1"
  [ -n "${2:-}" ] && printf '       %s\n' "$2"
}

# assert <description> <condition-exit-status> [detail]
assert() {
  if [ "$2" -eq 0 ]; then ok "$1"; else not_ok "$1" "${3:-}"; fi
}

# reset [<dns_records.json>] — put the stubs back into their default state.
reset() {
  cp "${1:-$STUB_DIR/dns_records.default.json}" "$STUB_DIR/dns_records.json"
  cp "$STUB_DIR/plan.json" "$WORK_DIR/tfplan"
  cp "$STUB_DIR/state.json" "$STUB_DIR/state.backup.json"
  : >"$STUB_DIR/calls.log"
  cd "$WORK_DIR" || exit 2
}

# run <mode> — run the script, capture its output, echo its exit status.
run() {
  "$SCRIPT_UNDER_TEST" "$@" >"$STUB_DIR/out.log" 2>&1
  echo $?
}

# called <pattern> — whether the stub Terraform saw a call matching it.
called() {
  grep -qF -- "$1" "$STUB_DIR/calls.log"
}

# imported <address> <zone/record id> — whether that import happened.
imported() {
  grep -qF -- "import -input=false $1 $2" "$STUB_DIR/calls.log"
}

printf '== the plan run only reports ==\n'
reset
rc="$(run --report)"
assert "the report run succeeds" "$([ "$rc" = 0 ] && echo 0 || echo 1)" "rc=$rc"
assert "the report run changes no state" "$(called 'state rm' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "the report run imports nothing" "$(called 'terraform import' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "the report run says what it would adopt" "$(grep -q 'would be replaced by the live A one' "$STUB_DIR/out.log" && echo 0 || echo 1)" "$(cat "$STUB_DIR/out.log")"
assert "the report run leaves the deeper name alone" "$(grep -q 'server.sub.opennhp.org' "$STUB_DIR/out.log" && echo 1 || echo 0)" "$(cat "$STUB_DIR/out.log")"

printf '== the apply run adopts what the state cannot describe ==\n'
reset
rc="$(run --apply)"
assert "the apply run succeeds" "$([ "$rc" = 0 ] && echo 0 || echo 1)" "rc=$rc out=$(cat "$STUB_DIR/out.log")"
assert "it removes the stale entry" "$(called 'state rm cloudflare_record.server' && echo 0 || echo 1)" "$(cat "$STUB_DIR/calls.log")"
assert "it imports the live A record" "$(imported cloudflare_record.server zone123/id-server-a && echo 0 || echo 1)" "$(cat "$STUB_DIR/calls.log")"
assert "it does not import the TXT record" "$(imported cloudflare_record.server zone123/id-server-txt && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it does not touch the records the state describes" "$(called 'state rm cloudflare_record.ac ' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it does not touch the retired name" "$(imported cloudflare_record.auth_plugin zone123/id-retired && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it plans again against the state it adopted" "$(called 'terraform plan' && echo 0 || echo 1)" "$(cat "$STUB_DIR/calls.log")"

printf '== a live record of another type is refused ==\n'
cat >"$STUB_DIR/dns_records.mismatch.json" <<'EOF'
{"success": true, "result": [
  {"id": "id-ac", "name": "ac.opennhp.org", "type": "A"},
  {"id": "id-server-cname", "name": "server.opennhp.org", "type": "CNAME"},
  {"id": "id-acdemo", "name": "acdemo.opennhp.org", "type": "CNAME"},
  {"id": "id-stats", "name": "stats.opennhp.org", "type": "A"}], "result_info": {"total_pages": 1}}
EOF
reset "$STUB_DIR/dns_records.mismatch.json"
rc="$(run --apply)"
assert "the apply run fails" "$([ "$rc" = 1 ] && echo 0 || echo 1)" "rc=$rc"
assert "it imports nothing" "$(called 'terraform import' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it removes nothing" "$(called 'state rm' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it says the state was not changed" "$(grep -q 'the state was not changed' "$STUB_DIR/out.log" && echo 0 || echo 1)" "$(cat "$STUB_DIR/out.log")"
assert "the report run only warns about it" "$(
  reset "$STUB_DIR/dns_records.mismatch.json"
  rc="$(run --report)"
  [ "$rc" = 0 ] && grep -q '::warning::' "$STUB_DIR/out.log"
  echo $?
)" "$(cat "$STUB_DIR/out.log")"

printf '== two live records of the declared type are refused ==\n'
cat >"$STUB_DIR/dns_records.two.json" <<'EOF'
{"success": true, "result": [
  {"id": "id-ac", "name": "ac.opennhp.org", "type": "A"},
  {"id": "id-server-a1", "name": "server.opennhp.org", "type": "A"},
  {"id": "id-server-a2", "name": "server.opennhp.org", "type": "A"},
  {"id": "id-acdemo", "name": "acdemo.opennhp.org", "type": "CNAME"},
  {"id": "id-stats", "name": "stats.opennhp.org", "type": "A"}], "result_info": {"total_pages": 1}}
EOF
reset "$STUB_DIR/dns_records.two.json"
rc="$(run --apply)"
assert "the apply run fails" "$([ "$rc" = 1 ] && echo 0 || echo 1)" "rc=$rc"
assert "it imports nothing" "$(called 'terraform import' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"

printf '== a state that cannot be read stops the run ==\n'
reset
MOCK_STATE_READ_FAIL=1
export MOCK_STATE_READ_FAIL
rc="$(run --apply)"
unset MOCK_STATE_READ_FAIL
assert "the apply run fails" "$([ "$rc" = 1 ] && echo 0 || echo 1)" "rc=$rc"
assert "it imports nothing" "$(called 'terraform import' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"
assert "it says the state could not be read" "$(grep -q 'could not read the Terraform state' "$STUB_DIR/out.log" && echo 0 || echo 1)" "$(cat "$STUB_DIR/out.log")"

printf '== a failed import puts the state back ==\n'
reset
MOCK_IMPORT_FAIL=1
export MOCK_IMPORT_FAIL
rc="$(run --apply)"
unset MOCK_IMPORT_FAIL
assert "the apply run fails" "$([ "$rc" = 1 ] && echo 0 || echo 1)" "rc=$rc"
assert "it restores the state from the backup" "$(called 'state push' && echo 0 || echo 1)" "$(cat "$STUB_DIR/calls.log")"
assert "it does not plan again" "$(called 'terraform plan' && echo 1 || echo 0)" "$(cat "$STUB_DIR/calls.log")"

printf '== every page of the zone is read ==\n'
cat >"$STUB_DIR/dns_records.page1.json" <<'EOF'
{"success": true, "result": [{"id": "id-ac", "name": "ac.opennhp.org", "type": "A"}], "result_info": {"total_pages": 2}}
EOF
cat >"$STUB_DIR/dns_records.page2.json" <<'EOF'
{"success": true, "result": [{"id": "id-page2", "name": "page2.opennhp.org", "type": "A"}], "result_info": {"total_pages": 2}}
EOF
cat >"$STUB_DIR/plan.page2.json" <<'EOF'
{"planned_values": {"root_module": {"resources": [
  {"address": "cloudflare_record.page2", "type": "cloudflare_record", "values": {"name": "page2", "type": "A"}}]}},
 "variables": {"domain": {"value": "opennhp.org"}}}
EOF
reset "$STUB_DIR/dns_records.page1.json"
cp "$STUB_DIR/plan.page2.json" "$WORK_DIR/tfplan"
rc="$(run --apply)"
assert "the apply run succeeds" "$([ "$rc" = 0 ] && echo 0 || echo 1)" "rc=$rc out=$(cat "$STUB_DIR/out.log")"
assert "it imports the record from the second page" "$(imported cloudflare_record.page2 zone123/id-page2 && echo 0 || echo 1)" "$(cat "$STUB_DIR/calls.log")"

printf '\n%s passed, %s failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
