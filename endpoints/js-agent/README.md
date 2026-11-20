# js-agent

Browser-side OpenNHP agent SDK (TypeScript). Mirrors the Go [`endpoints/agent`](../agent/)
client: it speaks the NHP protocol from a web page so a visitor's browser can
knock the demo NHP server and gain access to a protected resource without
installing a native binary.

This package was migrated from
[OpenNHP/js-agent](https://github.com/OpenNHP/js-agent) (subdirectory `nhp-js/`).
Only the SDK + the `relay-test.html` page used by the production demo pipeline
were brought over; the upstream repo's GitHub Pages landing page is not
included.

This directory is a standalone npm package and is **not** part of the Go
workspace defined by [`endpoints/go.mod`](../go.mod) — `go build ./...` ignores
it.

## Layout

```text
src/
  NHPAgent.ts          high-level agent API
  index.ts, types.ts   public exports
  crypto/              X25519 / SM2 / AES-GCM / SM4 / BLAKE2s / SM3
  protocol/            NHP packet header + framing
  transport/           UDP, WebSocket, HTTP relay
test/                  vitest unit tests for SDK + crypto + protocol
examples/
  relay-test.html      browser demo page (deployed at https://agent.opennhp.org/)
```

## Build

```bash
npm ci
npm run build      # emits dist/index.js + dist/index.d.ts (vite + tsc)
npm run test:run   # vitest
```

## Version

The SDK's version number comes from `nhp/version/VERSION`, the same file the Go
binaries read. It is injected at build time by `vite.config.ts` as
`__SDK_VERSION__` and exposed as `NHPAgent.version`. The public pages render it
(agent.opennhp.org, reg.opennhp.org and demo.opennhp.org), so it must match the
released version.

It is kept in sync automatically. `release-please-config.json` declares
`nhp/version/VERSION` as the release's `version-file` and this package's
`package.json` (plus `package-lock.json`) under `extra-files`, so the
release-please release PR bumps them together with
`.release-please-manifest.json`. `scripts/check-version-sync.sh` (run in CI and
via `make check-version-sync`) fails the build if they ever drift apart.

> The `release-please` workflow deliberately passes **no** `release-type` input
> to the action: the action only reads `release-please-config.json` when that
> input is absent. With it set, the action runs in "simple" mode and ignores
> the config file, so the version files would silently stop being updated.

## Demo deployment

Built and deployed by the `deploy-jsagent` job in
[.github/workflows/deploy-demo-v2.yml](../../.github/workflows/deploy-demo-v2.yml):

1. `npm ci && npm run build` in this directory.
2. `examples/relay-test.html` is rewritten so its `import` resolves to
   `./nhp-js/dist/index.js` and served as `index.html`.
3. `dist/` is copied alongside it and a `config.json` is rendered with the
   public demo agent key pair from AWS Secrets Manager (`opennhp/demo`,
   field `nhp_jsagent_*`).
4. The bundle is rsync'd to `/var/www/jsagent/` on the relay host and served
   over TLS by the nginx vhost in
   [deploy/nginx/jsagent.conf.template](../../deploy/nginx/jsagent.conf.template).

The `agentPrivateKey` exposed in `config.json` is intentionally public — see
the comment in the workflow before reusing it for anything else.

## Runtime dependency on esm.sh

The browser demo loads the `@noble/{ciphers,curves,hashes}` packages from
[esm.sh](https://esm.sh/) at runtime via the `<script type="importmap">` block
in `examples/relay-test.html`. They are declared as `external` in
[vite.config.ts](vite.config.ts) and therefore are *not* bundled into
`dist/index.js`. This keeps the SDK build small but means the demo page
depends on a third-party CDN being reachable. If you need a self-contained
deployment, either remove the `external` entries (bundle the noble libs) or
copy them next to `dist/` and point the importmap at local paths.

> **Security note** — pulling cryptographic primitives over a third-party CDN
> at runtime means the demo's confidentiality relies on esm.sh (and any
> intermediary it uses) not being compromised or impersonated. The demo's
> `agentPrivateKey` is published anyway, so this is acceptable for the demo
> page. **Anyone reusing this template for a non-demo deployment must bundle
> the `@noble/*` modules locally** — do not ship a real agent that imports
> crypto code from a CDN you don't control. By contrast, `sm-crypto-v2` is
> *not* externalized and is bundled into `dist/index.js`; consider similarly
> bundling all crypto dependencies for production use.
