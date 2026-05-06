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

## Layout

```text
src/
  NHPAgent.ts          high-level agent API
  index.ts, types.ts   public exports
  crypto/              X25519 / SM2 / AES-GCM / SM4 / BLAKE2s / SM3
  protocol/            NHP packet header + framing
  transport/           UDP, WebSocket, WebRTC, HTTP relay
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
