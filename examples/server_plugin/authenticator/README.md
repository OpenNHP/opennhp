# Authenticator server plugin

An nhp-server plugin that opens a protected resource after the user proves
possession of a TOTP code (Google Authenticator, Authy, ...). It offers two
ways to log in, and both check codes generated from the single configured
`OTPSecretKey`:

- **Manual OTP**: the user types the 6-digit code into the login page.
- **QR login**: the login page shows a QR code, a phone scans it and submits a
  code, and the browser then completes the knock.

All endpoints are served by nhp-server at
`/plugins/authenticator?resid=<ResourceId>&action=<action>`.

## Build

```bash
make            # builds authenticator.so into ../../../release/nhp-server/plugins/authenticator/
go test ./...   # unit and handler tests
```

## Configuration (`etc/config.toml`)

| Key | Default | Meaning |
| --- | --- | --- |
| `OTPSecretKey` | — (required) | Base32 TOTP seed shared by every user. Enrol it in each user's authenticator app. |
| `OTPPeriod` | `30` | TOTP step in seconds, 15–300. Out-of-range values fall back to 30. |
| `OTPDigits` | `6` | Only 6 is supported (the login page takes six digits); other values fall back to 6 with an error log. |
| `QRCodeExpiry` | `300` | QR session lifetime in seconds. |
| `OTPAllowReuse` | `false` | When `false`, a code that was accepted once is refused for its remaining validity (about 90 s). |
| `ShowOTPSecret` | `false` | Render `OTPSecretKey` on the public login page so visitors can enrol it. **Public demos only**: anyone who can load the page can then generate valid codes. |
| `TrustedProxies` | empty | Reverse proxies in front of nhp-server (addresses or CIDR prefixes). **Must be set behind a proxy**; see below. IPv4-mapped entries (`::ffff:10.0.0.0/104`) are treated as IPv4; an IPv4-mapped prefix shorter than /96 is invalid. One invalid entry disables the whole list. |

The algorithm is fixed at TOTP / SHA1 / 6 digits, accepting one step of clock
skew either side of the current step.

### Shared seed: what operators should know

Every user shares one `OTPSecretKey`. That has two consequences:

- **With `OTPAllowReuse = false`, only about three people can log in per
  90-second window.** Everyone else gets "OTP code already used" until the next
  code. Multi-user deployments should set `OTPAllowReuse = true`.
- **The global rate limit is the only bound on a distributed attacker, and it
  fails closed.** See *Rate limiting* below.

## Rate limiting

Every code check, on both `otpvalid` and the QR flow's `verify`, goes through
one limiter:

- **Per address**: 5 failures lock the address out for 15 minutes. IPv6
  clients are keyed by /64, and each /48 additionally gets at most 15 failures,
  so rotating addresses inside one allocation does not help.
- **Global**: at most 30 failures in any minute (sliding window), across all
  addresses.
- **Per QR session**: at most 5 wrong codes, after which the session is dead
  and a new QR code is needed.
- An attempt is counted *before* the code is checked and given back if the code
  was correct, so parallel requests cannot all slip in before a failure is
  recorded. A correct code refused only because it was already used is not
  counted as a failure.

The global limit fails closed: while it is reached, OTP and QR login are blocked
for everyone, and `[OTP] global limit reached` is logged once per minute.
Because of the per-address lockouts, keeping it reached takes about 90 IPv4
addresses (or 30 IPv6 /48s) every 15 minutes. That is cheap for a botnet but not
possible from a single host. The global limit also bounds a distributed attacker
to roughly a 12% chance of a correct guess per day (about a week expected) while
`ShowOTPSecret` is off.

### Client address and reverse proxies

The limiter and the knock both use the client's address, never a value the
client can forge:

- By default that is the TCP peer. `X-Forwarded-For` is ignored.
- If the peer address cannot be parsed (for example a unix-socket listener in
  front of nhp-server), the request is refused with HTTP 400, rather than
  sharing one limit and opening access for an empty address.
- If the peer is listed in `TrustedProxies`, it is the right-most
  `X-Forwarded-For` hop that is not itself a trusted proxy.

Access is opened for that same address. The plugin overwrites the knock's
source address, which nhp-server would otherwise take from `X-Forwarded-For`
for any peer. **Behind a reverse proxy, set `TrustedProxies`.** Otherwise every
client shares the proxy's limit, one client can lock out everyone, and access is
opened for the proxy's address instead of the client's. A warning is logged the
first time a loopback or private peer sends `X-Forwarded-For` while
`TrustedProxies` is empty.

## Manual OTP login

```
GET ?resid=<res>&action=otpvalid&otpCode=<6 digits>
```

On success the knock is performed and the nhp-server ack is returned
(`errCode` `"0"`, optional `redirectUrl`). Failures return
`{"success": false, "errMsg": ...}`:

| HTTP | `errMsg` |
| --- | --- |
| 200 | `Invalid OTP code` |
| 200 | `OTP code already used, wait for the next code` |
| 429 | `Too many failed attempts, try again later` |
| 400 | `Cannot determine the client address` |

The code is marked used before the knock is sent, so a knock that fails on the
server or AC side (for example, an unreachable AC) still spends it. With
`OTPAllowReuse = false`, wait for the next code before retrying.

## QR login protocol

```
browser                         nhp-server (plugin)                    phone
   | action=generate  ───────────▶ new session                           |
   | ◀─── sessionId, qrData        (no secret is returned)               |
   | render nhp://scan?...  ─────────────── scan ──────────────────────▶ |
   |                               ◀── action=scan (optional) ───────────|
   |                               ◀── action=verify(qrData, otpCode) ───|
   |                                   code checked against OTPSecretKey |
   | action=status (poll) ───────▶ status = confirmed                    |
   | action=qrvalid ─────────────▶ knock (once, same client, same res)   |
```

### `action=generate` (browser)

Response: `success`, `sessionId`, `qrData` (encrypted, signed session data),
`serverUrl`, `aspId`, `resId`, `expiresAt` (Unix seconds).

`generate` is unauthenticated, so sessions are bounded: at most 20 per client
(the same limit key as the rate limiter) and 10,000 in total. A session holds its
slot until it is cleaned up, about ten minutes after `generate`. Beyond that,
the response is HTTP 429 `too many QR codes generated, try again later`.

The login page encodes the QR code as:

```
nhp://scan?data=<qrData>&server=<serverUrl>&asp=<aspId>&res=<resId>&sid=<sessionId>
```

### `action=scan` (phone, optional)

`?sessionId=<sid>`. Marks a pending session as scanned, for the browser's UI.
The request is unauthenticated, so it only ever moves a session from pending to
scanned. It never changes a confirmed, failed or expired session.

### `action=verify` (phone)

`GET ?resid=<res>&action=verify&encryptedData=<data>&otpCode=<code>&deviceInfo=<text>`,
or `POST` the same fields as JSON: `{"encryptedData", "otpCode", "deviceInfo"}`.

`otpCode` must be the **current TOTP code for the configured `OTPSecretKey`**
(SHA1, 6 digits, step `OTPPeriod`), so the phone must have that secret enrolled.
Response: `{"success": true}` or `{"success": false, "errMsg": ...}`:

| HTTP | `errMsg` | Suggested client behaviour |
| --- | --- | --- |
| 200 | `invalid OTP code` | Let the user retry (at most 5 per session). |
| 200 | `OTP code already used` | Wait for the next code. |
| 200 | `too many failed attempts, generate a new QR code` | Ask the user to refresh the QR code on the computer. |
| 200 | `session not found` / `session expired` / `session already confirmed` / `invalid QR data` | Ask the user to refresh the QR code. |
| 429 | `too many failed OTP attempts` | Rate limited; try again later. |
| 400 | `cannot determine the client address` | Server-side setup problem; report it to the operator. |

### `action=status` (browser)

`?sessionId=<sid>`. Returns `status`: `0` pending, `1` scanned, `2` confirmed,
`3` expired, `4` failed (a wrong code was entered; the phone may retry).

### `action=qrvalid` (browser)

`?sessionId=<sid>`. Performs the knock only if the session is confirmed and
unexpired, belongs to `resid`, has not knocked before, and the request comes
from the same client address (same limiter key) that called `generate`. If the
browser's address changed in between, for example a dual-stack browser
switching between IPv4 and IPv6, it gets "this QR code was generated from a
different network address; refresh the QR code and try again". Changing
`TrustedProxies` while a QR login is in progress can have the same effect,
because `qrvalid` derives the requester's address with the current list; a new
QR code fixes it.

## Breaking change: QR protocol (per-session secret removed)

Earlier versions created a TOTP secret for every QR session. `generate`
returned it as `otpSecret` / `otpUri`, the QR code carried it as `otp=`, and
`verify` checked the phone's code against it.

That design authenticated nothing. The QR code is rendered by the browser, so
whoever requested it also held the secret. Any client could call `generate`,
compute the code itself, call `verify` and then `qrvalid`, and get access
without knowing `OTPSecretKey` and without ever hitting the rate limit. The
phone has to prove something the browser cannot know, and nothing carried in
the QR code can be that. So the fix cannot keep the old protocol:

| | Before | Now |
| --- | --- | --- |
| `generate` response | included `otpSecret`, `otpUri` | no secret |
| QR payload | `nhp://scan?data=…&otp=<secret>&…` | no `otp=` |
| `verify` checks the code against | the per-session secret | `OTPSecretKey`, under the shared rate limit |
| `qrvalid` | knocked for any non-empty `sessionId` | confirmed session, same client, same resource, once |

### Impact

- **Clients in this repository are not affected.** The Android and iOS samples
  under `examples/client_sdk/` only initialise the nhp-agent SDK and run the
  NHP knock loop over UDP. They never call this plugin's HTTP endpoints. No
  code in this repository consumes `nhp://scan`.
- **External mobile clients that derived the code from `otp=` will fail
  `verify` on every attempt** against an upgraded server, so QR login stops
  working for their users until the client is updated.

### Migrating a mobile client

1. Do not require `otp` when parsing `nhp://scan`. Keep reading `data`,
   `server`, `asp`, `res` and `sid`.
2. Compute the code from the user's enrolled `OTPSecretKey`
   (TOTP / SHA1 / 6 digits / `OTPPeriod`), or have the user type it from their
   authenticator app. If the app generated codes itself from `otp=`, it now
   needs a step that enrols `OTPSecretKey`, distributed by the operator.
3. Send `verify` exactly as before (`encryptedData`, `otpCode`, `deviceInfo`),
   and handle the responses in the table above, including HTTP 429.
4. To work against servers on both sides of the upgrade, use `otp=` when the
   QR code carries it (old server) and the enrolled `OTPSecretKey` when it does
   not (new server). The client can then ship before the server is upgraded.

Operators: upgrading the server before every user's QR client supports the new
protocol breaks QR login for those users. Manual OTP login works throughout.
