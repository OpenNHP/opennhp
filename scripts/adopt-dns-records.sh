#!/usr/bin/env bash
# adopt-dns-records.sh
#
# Adopt the Cloudflare DNS records that exist in the zone but that the
# Terraform state cannot describe.
#
# An apply that creates a DNS record and then fails before the provider
# records its id leaves the record in Cloudflare and out of the state: the
# next plan tries to create that record again, and the provider refuses it
# -- "expected DNS record to not already be present but already exists" --
# so the pipeline cannot get out of that on its own. This script repairs
# that: it imports the live record into the state, so the plan that follows
# sees a record that is already there and leaves it alone.
#
# The repair is deliberately narrow. It happens only where the state cannot
# describe the live record at all:
#
#   * the record is missing from the state, or
#   * the state holds it as a record of another type. The provider cannot
#     change a record's type in place -- it has to delete the record and
#     create it again, and it refuses to create a record whose name is
#     taken -- so a type change takes two applies, and the name does not
#     resolve in between, or
#   * the state holds it under another id: the record the state names is
#     not the record the zone holds. That happens when the record was
#     deleted and created again in the dashboard, or when an apply
#     destroyed it and created it again before it died. The state then
#     cannot describe the live record at all, and the next plan fails with
#     "expected DNS record to not already be present but already exists"
#     -- the very failure this script exists to repair.
#
# Where the state and Cloudflare agree on the type *and* the id, nothing is
# touched and Terraform's normal drift detection applies, so a record whose
# content was changed out of band is still corrected by the plan. The
# adoption is a repair, not a drift detector.
#
# Usage:
#   scripts/adopt-dns-records.sh --report    # read-only: report what would be adopted
#   scripts/adopt-dns-records.sh --apply     # import what the state cannot describe
#
# --report changes nothing: it reads the plan, the state and the zone, and
# prints what an --apply run would do. It is what a plan run uses, because a
# plan run must not touch the state -- it runs unattended on every push to
# main and needs no confirmation.
#
# --apply imports the records and re-plans, so the plan that gets applied is
# the one that matches the state it adopted. It refuses to adopt a record
# whose type the configuration does not declare: the plan that follows would
# have to delete that record and create it again, which is the very window
# that leaves a name unresolvable. That is a decision for a human.
#
# Run it in the Terraform working directory (terraform/demo): it reads the
# plan file and the remote state from there, and writes the state back.
#
# Environment:
#   CF_TOKEN   Cloudflare API token (Zone:DNS:Read, Zone:Zone:Read)
#   CF_ZONE    the Cloudflare zone id
#   PLAN_FILE  the plan to read (default: tfplan)
#
# The token is never put on a command line; it reaches curl through its
# configuration on stdin (see "the Cloudflare API" below).
#
# Requires: terraform, jq, curl.
#
# Tested by tests/scripts/test_adopt_dns_records.sh, which runs it against a
# stubbed Terraform and a stubbed Cloudflare API.

set -euo pipefail

MODE=""
PLAN_FILE="${PLAN_FILE:-tfplan}"

usage() {
  cat <<'EOF'
Usage: scripts/adopt-dns-records.sh (--report | --apply)

  --report   Read-only. Report which live records the state cannot describe.
  --apply    Import those records into the state, then re-plan.

Environment:
  CF_TOKEN   Cloudflare API token (Zone:DNS:Read, Zone:Zone:Read)
  CF_ZONE    the Cloudflare zone id
  PLAN_FILE  the plan file to read (default: tfplan)
EOF
}

note() { printf '::notice::%s\n' "$*"; }
warn() { printf '::warning::%s\n' "$*"; }
err() { printf '::error::%s\n' "${*#ERROR: }"; }
die() { err "$*"; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --report) MODE="report" ;;
    --apply) MODE="apply" ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      printf 'Unknown option: %s\n\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

if [ -z "$MODE" ]; then
  printf 'ERROR: exactly one of --report or --apply is required.\n\n' >&2
  usage >&2
  exit 2
fi

# --- preconditions --------------------------------------------------------

for cmd in terraform jq curl; do
  command -v "$cmd" >/dev/null 2>&1 || die "ERROR: '$cmd' is required but was not found on PATH."
done
[ -n "${CF_TOKEN:-}" ] || die "ERROR: CF_TOKEN is not set."
[ -n "${CF_ZONE:-}" ] || die "ERROR: CF_ZONE is not set."
[ -d ".terraform" ] || die "ERROR: no .terraform directory here; run this in the Terraform working directory (terraform/demo)."
[ -f "$PLAN_FILE" ] || die "ERROR: the plan file '${PLAN_FILE}' does not exist; run 'terraform plan -out=${PLAN_FILE}' first."

plan_json="$(mktemp)"
state_json="$(mktemp)"
state_list="$(mktemp)"
decisions="$(mktemp)"
refusal_file="$(mktemp)"
backup_file="$(mktemp)"
stderr_file="$(mktemp)"
trap 'rm -f "$plan_json" "$state_json" "$state_list" "$decisions" "$refusal_file" "$backup_file" "$stderr_file"' EXIT

# --- the Cloudflare API ----------------------------------------------------
#
# The token is never put on a command line: the arguments of a process are
# readable by every other process on the machine (ps, /proc/<pid>/cmdline),
# and this script runs unattended on the CI runner on every push. curl
# reads its configuration from stdin with "--config -", so the header
# travels that way and never appears in argv.
#
# curl's configuration file reads a double-quoted value with the usual
# backslash escapes, so the token is escaped before it goes in.

cf_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  printf '%s' "$s"
}

cf_get() {
  printf 'header = "Authorization: Bearer %s"\n' "$(cf_escape "$CF_TOKEN")" |
    curl -sS --retry 3 --config - "$1"
}

# --- the plan -------------------------------------------------------------
#
# The configuration's intent: one cloudflare_record per hostname, with the
# name and the record type it asks for.

if ! terraform show -json "$PLAN_FILE" >"$plan_json" 2>"$stderr_file"; then
  die "ERROR: could not read the plan file '${PLAN_FILE}': $(cat "$stderr_file")"
fi

records="$(jq -r '
  def mod_resources:
    (.resources // [])[], ((.child_modules // [])[] | mod_resources);
  [ .planned_values.root_module
    | mod_resources
    | select(.type == "cloudflare_record")
    | [ .address, (.values.name // ""), (.values.type // "") ]
    | join("|")
  ] | .[]' "$plan_json")"

cfg_domain="$(jq -r '.variables.domain.value // empty' "$plan_json")"

# --- the zone -------------------------------------------------------------

if ! zone_json="$(cf_get "https://api.cloudflare.com/client/v4/zones/${CF_ZONE}")"; then
  die "ERROR: could not read the Cloudflare zone ${CF_ZONE}."
fi
if ! jq -e '.success == true' >/dev/null <<<"$zone_json"; then
  die "ERROR: Cloudflare refused to describe the zone ${CF_ZONE}: $(jq -r '.errors[0].message // "unknown error"' <<<"$zone_json")"
fi
zone_name="$(jq -r '.result.name // empty' <<<"$zone_json")"
[ -n "$zone_name" ] || die "ERROR: the Cloudflare zone ${CF_ZONE} has no name."

if [ -n "$cfg_domain" ] && [ "$cfg_domain" != "$zone_name" ]; then
  warn "The configuration's domain (${cfg_domain}) is not the zone's name (${zone_name}); the live records are matched against the zone's name."
fi

# --- the live records -----------------------------------------------------
#
# Every record the zone holds, fetched once. The API's own name filter is a
# substring match, so the records are matched here instead, on the whole
# name and the record type.

live_records='[]'
page=1
while :; do
  if ! page_json="$(cf_get "https://api.cloudflare.com/client/v4/zones/${CF_ZONE}/dns_records?per_page=100&page=${page}")"; then
    die "ERROR: could not list the DNS records of the zone ${CF_ZONE}."
  fi
  if ! jq -e '.success == true' >/dev/null <<<"$page_json"; then
    die "ERROR: Cloudflare refused to list the DNS records of the zone ${CF_ZONE}: $(jq -r '.errors[0].message // "unknown error"' <<<"$page_json")"
  fi
  live_records="$(jq -c --argjson acc "$live_records" '$acc + (.result // [])' <<<"$page_json")"
  total_pages="$(jq -r '.result_info.total_pages // 1' <<<"$page_json")"
  if [ "$page" -ge "$total_pages" ]; then
    break
  fi
  page=$((page + 1))
done

# --- the state ------------------------------------------------------------
#
# A state that cannot be read is not a state without the record: it is a
# state this script does not know how to reason about, and treating it as
# "the record is absent" is how a record gets removed from the state by
# accident. Both reads below fail the run instead.

if ! terraform state list >"$state_list" 2>"$stderr_file"; then
  die "ERROR: could not read the Terraform state: $(cat "$stderr_file")"
fi
if ! terraform show -json >"$state_json" 2>"$stderr_file"; then
  die "ERROR: could not read the Terraform state: $(cat "$stderr_file")"
fi
if ! jq -e '.values.root_module' "$state_json" >/dev/null 2>&1; then
  die "ERROR: the state JSON has no root module; refusing to guess what the state holds."
fi

# What the state holds for an address: one of the record's fields. The type
# says whether the state describes the live record at all; the id says
# whether it describes *this* record -- a state that names an id the zone
# does not hold describes nothing that is live.
state_field_of() {
  jq -r --arg a "$1" --arg f "$2" '
    def mod_resources:
      (.resources // [])[], ((.child_modules // [])[] | mod_resources);
    [ .values.root_module
      | mod_resources
      | select(.address == $a)
      | (.values[$f] // empty)
    ] | .[0] // empty' "$state_json"
}

state_type_of() { state_field_of "$1" type; }
state_id_of() { state_field_of "$1" id; }

# --- analyse --------------------------------------------------------------
#
# Every decision is taken before anything is changed, so a record that needs
# a human decision stops the run before the state is touched at all.

: >"$decisions"
: >"$refusal_file"
refusals=0

# One decision per line: what the plan declares, what the zone holds, what
# the state holds, and what is done about the difference.
#
#   address|fqdn|declared type|live id|live type|state type|state id|action
#
decide() {
  printf '%s|%s|%s|%s|%s|%s|%s|%s\n' \
    "$1" "$2" "$3" "$4" "$5" "$6" "$7" "$8" >>"$decisions"
}

# A refusal is recorded here and reported after the table, so the run reads
# as one report rather than as a mix of the two.
refuse() {
  local addr="$1" fqdn="$2" want="$3" live_type="$4" state_type="$5" state_id="$6" msg="$7"
  refusals=$((refusals + 1))
  decide "$addr" "$fqdn" "$want" "" "$live_type" "$state_type" "$state_id" "refuse"
  printf '%s\n' "$msg" >>"$refusal_file"
}

while IFS='|' read -r addr name want; do
  [ -n "$addr" ] || continue

  if [ -z "$name" ] || [ -z "$want" ]; then
    warn "The plan does not give ${addr} a name and a type; leaving it alone."
    continue
  fi

  # Cloudflare answers with the fully qualified name ("server.opennhp.org")
  # while the configuration names the record relative to the zone ("server").
  # The whole name is compared: matching the first label alone would also
  # match every deeper name, and one name can hold more than one record.
  fqdn="$name"
  case "$fqdn" in
  "${zone_name}") ;;                     # the zone's own name
  *."${zone_name}") ;;                   # already fully qualified
  *) fqdn="${name}.${zone_name}" ;;
  esac

  # What the state holds for this address, if it holds it at all. The
  # presence comes from "terraform state list" and the content from the
  # machine-readable state; both were read above, and a failure to read
  # either of them has already stopped the run.
  state_type="$(state_type_of "$addr")"
  state_id="$(state_id_of "$addr")"
  if grep -Fxq -- "$addr" "$state_list"; then
    state_present="yes"
  else
    state_present="no"
  fi

  candidates="$(jq -c --arg fqdn "$fqdn" \
    '[.[] | select(((.name // "") | ascii_downcase) == ($fqdn | ascii_downcase))]' \
    <<<"$live_records")"
  candidate_count="$(jq 'length' <<<"$candidates")"

  if [ "$candidate_count" -eq 0 ]; then
    # Nothing is live under that name: Terraform creates it.
    decide "$addr" "$fqdn" "$want" "" "" "$state_type" "$state_id" "create"
    continue
  fi

  matched="$(jq -c --arg want "$want" '[.[] | select((.type // "") == $want)]' <<<"$candidates")"
  matched_count="$(jq 'length' <<<"$matched")"

  if [ "$matched_count" -gt 1 ]; then
    live_types="$(jq -r '[.[].type] | unique | join(", ")' <<<"$candidates")"
    refuse "$addr" "$fqdn" "$want" "$live_types" "$state_type" "$state_id" \
      "The zone holds ${matched_count} ${want} records for ${fqdn}. Refusing to choose one of them by guesswork: the wrong one would be adopted and then replaced by the next apply. Delete the extra records in Cloudflare, or give the configuration a name of its own."
    continue
  fi

  if [ "$matched_count" -eq 0 ]; then
    live_types="$(jq -r '[.[].type] | unique | join(", ")' <<<"$candidates")"
    refuse "$addr" "$fqdn" "$want" "$live_types" "$state_type" "$state_id" \
      "The live record for ${fqdn} is a ${live_types} but the configuration declares ${want}. Adopting it would leave the plan with a record type to change, which the Cloudflare provider cannot do in place: it has to delete the record and create it again, and a record of that name is already there -- the window in which the name does not resolve. Delete the live record in Cloudflare first, or change the configuration to declare the type that is live."
    continue
  fi

  live_id="$(jq -r '.[0].id // empty' <<<"$matched")"
  live_type="$(jq -r '.[0].type // empty' <<<"$matched")"
  if [ -z "$live_id" ]; then
    refuse "$addr" "$fqdn" "$want" "$live_type" "$state_type" "$state_id" \
      "The live ${want} record for ${fqdn} has no id; there is nothing to import."
    continue
  fi

  if [ "$state_present" != "yes" ]; then
    action="adopt"
  elif [ -z "$state_type" ] || [ "$state_type" != "$live_type" ]; then
    # The state holds the address as a record of another type, or as a
    # record it cannot describe: the provider cannot change a record's
    # type in place, so the entry has to go before the live record can be
    # imported under that address.
    action="replace"
  elif [ -n "$state_id" ] && [ "$state_id" != "$live_id" ]; then
    # The state holds the address under another id: the record it names is
    # not the one the zone holds. The record was deleted and created again
    # in the dashboard, or an apply destroyed it and created it again
    # before it died, and the state now describes nothing that is live.
    # The plan would refresh the id, get a 404, drop the entry and plan a
    # create -- which fails with "expected DNS record to not already be
    # present but already exists". The entry goes and the live record is
    # imported in its place.
    #
    # A state that names no id at all is left alone: it cannot be compared
    # with anything, and guessing that it describes a different record
    # would rewrite the whole state on every run.
    action="replace"
  else
    # The state describes the live record; leave it to Terraform.
    action="insync"
  fi

  decide "$addr" "$fqdn" "$want" "$live_id" "$live_type" "$state_type" "$state_id" "$action"
done <<<"$records"

# --- report ---------------------------------------------------------------

printf 'Cloudflare DNS records vs the Terraform state (zone %s, %s):\n' "$CF_ZONE" "$zone_name"
while IFS='|' read -r addr fqdn want live_id live_type state_type state_id action; do
  [ -n "$addr" ] || continue
  case "$action" in
  create) verdict="not in Cloudflare; Terraform will create it" ;;
  insync) verdict="the state describes the live record" ;;
  adopt) verdict="the state has no record for it: it would be imported" ;;
  replace)
    if [ -z "$state_type" ] || [ "$state_type" != "$live_type" ]; then
      verdict="the state holds a ${state_type:-record of an unknown type} where the zone holds a ${live_type}: it would be replaced by the live ${live_type} one"
    else
      verdict="the state holds the id ${state_id:-none} and the zone holds ${live_id}: it would be replaced by the live ${live_type} one"
    fi
    ;;
  refuse) verdict="needs a decision from a human" ;;
  *) verdict="?" ;;
  esac
  printf '  %-26s %-26s live %-6s state %-6s %s\n' \
    "$addr" "$fqdn" "${live_type:--}" "${state_type:-absent}" "$verdict"
done <"$decisions"

if [ "$refusals" -gt 0 ]; then
  while IFS= read -r msg; do
    [ -n "$msg" ] || continue
    if [ "$MODE" = "apply" ]; then
      err "$msg"
    else
      warn "$msg"
    fi
  done <"$refusal_file"
fi

if [ "$MODE" != "apply" ]; then
  if [ "$refusals" -gt 0 ]; then
    warn "${refusals} record(s) need a decision from a human; an apply would stop there, with the state unchanged."
  fi
  exit 0
fi

# --- adopt ----------------------------------------------------------------

if [ "$refusals" -gt 0 ]; then
  die "ERROR: ${refusals} record(s) need a decision from a human. Nothing was adopted and the state was not changed."
fi

backup_taken=0

take_state_backup() {
  if [ "$backup_taken" -eq 1 ]; then
    return 0
  fi
  if ! terraform state pull >"$backup_file" 2>"$stderr_file"; then
    die "ERROR: could not back the Terraform state up; refusing to change it."
  fi
  backup_taken=1
}

restore_state_backup() {
  if [ "$backup_taken" -ne 1 ]; then
    return 0
  fi
  if terraform state push -force "$backup_file" >/dev/null 2>"$stderr_file"; then
    err "The Terraform state was restored from the backup taken before the adoption."
  else
    err "The Terraform state could not be restored automatically: $(cat "$stderr_file")"
    err "Restore it by hand with: terraform state push -force ${backup_file}"
  fi
}

adopted=0
while IFS='|' read -r addr fqdn want live_id live_type state_type state_id action; do
  [ -n "$addr" ] || continue
  case "$action" in
  adopt | replace) ;;
  *) continue ;;
  esac

  note "Adopting the live ${fqdn} record (${live_type}, id ${live_id}) into the state as ${addr} (state: ${state_type:-absent}${state_id:+, id ${state_id}}, configuration wants: ${want})."

  if [ "$action" = "replace" ]; then
    # The state holds the address, but not as this record: as a record of
    # another type, or under an id the zone does not hold. The import
    # cannot write over it, so the entry goes first -- and the state is
    # backed up before it does, so a failed import can be undone instead of
    # leaving the record in Cloudflare and out of the state, which is the
    # very orphan this script exists to repair.
    take_state_backup
    if ! terraform state rm "$addr"; then
      err "ERROR: could not remove ${addr} from the state; the live record was not adopted."
      restore_state_backup
      exit 1
    fi
  fi

  # The provider's importer takes "<zoneID>/<recordID>" and refuses a bare
  # record id: "invalid id ... should be in format \"zoneID/recordID\" for
  # import".
  if ! terraform import -input=false "$addr" "${CF_ZONE}/${live_id}"; then
    err "ERROR: the import of ${addr} failed."
    restore_state_backup
    exit 1
  fi
  adopted=$((adopted + 1))
done <"$decisions"

if [ "$adopted" -gt 0 ]; then
  # The plan was computed against the state as it was before the adoption,
  # so it still asks for the record to be created. Plan again against the
  # adopted state: the plan that gets applied has to be the one that
  # matches the state.
  note "Adopted ${adopted} record(s); planning again against the state it adopted."
  if ! terraform plan -input=false -out="$PLAN_FILE"; then
    die "ERROR: the re-plan failed after the adoption."
  fi
fi
