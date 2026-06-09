# Local multi-cluster validation stack

Two **independent** server clusters behind one shared relay, used to validate
multi-cluster routing locally before the AWS demo deploy (server cluster 2).

| Role | Container | Internal IP | Cluster |
| --- | --- | --- | --- |
| nhp-server | `nhp-server` | 177.7.0.9 | 1 |
| nhp-ac | `nhp-ac` | 177.7.0.10 | 1 |
| nhp-server-c2 | `nhp-server-c2` | 177.7.0.20 | 2 |
| nhp-ac-c2 | `nhp-ac-c2` | 177.7.0.21 | 2 |
| nhp-relay (shared) | `nhp-relay` | 177.7.0.12 → host **:8081** | 1 + 2 |
| web-app (protected) | `web-app` | 177.7.0.11:80 | shared |

Clusters 1 and 2 use **independent GMSM keypairs** and **separate ACs**
(`testAC-1` vs `testAC-c2`). The relay declares both as distinct `[[Servers]]`
entries, so `GET /servers` returns two fingerprints and
`POST /relay/<fingerprint>` routes to the chosen cluster.

Configs live under `multicluster/` (server-c2, ac-c2, relay). The stack runs
the `:current` images (built from current source — see "Rebuilding images").

## Start / restart

```bash
cd docker
docker compose -f docker-compose.multicluster.yaml up -d --no-build
docker compose -f docker-compose.multicluster.yaml ps          # 6 containers running
```

`--no-build` uses the prebuilt `:current` images and avoids the network-heavy
image build.

## Verify

### A. Relay exposes both clusters

```bash
curl -s http://127.0.0.1:8081/servers | python3 -m json.tool
```

Expect two entries: `demo` and `demo-cluster2`, each with a distinct `id`
(public-key fingerprint) and `publicKeyBase64`.

### B. Knock cluster 2 (success + AC token)

```bash
node multicluster/knock-tests/knock-cluster2.mjs
```

Expect `"success": true` with `resourceHosts` and an AC token. The SDK derives
the relay path `/relay/Stom6uqf5mM` from cluster 2's server public key.

### C. Confirm cluster 2 AC opened the firewall (within ~15s of the knock)

```bash
node multicluster/knock-tests/knock-cluster2.mjs >/dev/null
docker exec nhp-ac-c2 sh -c 'ipset list defaultset | sed -n "/Members/,\$p"'
```

Expect live entries like `192.168.65.1,tcp:80,177.7.0.11 timeout 19` — the
cluster 2 AC (`testAC-c2`) allowing the agent to reach the protected web-app.

### D. Cluster 1 still works independently (isolation)

```bash
node multicluster/knock-tests/knock-cluster1.mjs
```

Expect `"success": true` — cluster 1 knocks via `/relay/jf0Ied5gmZQ`, unaffected
by cluster 2.

## Tear down

```bash
docker compose -f docker-compose.multicluster.yaml down
```

## Rebuilding images (`:current`)

The committed prebuilt images can be stale (predating the `[[Servers]]` relay/AC
schema). Rebuild current-source images **offline** (the public Go proxy is often
flaky) by compiling inside the base image with the host module cache mounted,
then overlaying onto the runtime images:

```bash
cd <repo-root>
# 1. Build serverd + acd + plugins inside linux (CGO needed for plugins) using
#    the host's warm module cache, no network.
docker run --rm -v "$PWD":/workdir -w /workdir \
  -v "$(go env GOMODCACHE)":/go/pkg/mod -e GOPROXY=off -e GOFLAGS=-mod=mod \
  --entrypoint sh opennhp-base:latest -c 'make serverd acd plugins'

# 2. Overlay fresh binaries/plugins onto the existing runtime images.
docker build -t opennhp-server:current -f - . <<'DOCKER'
FROM opennhp-server:latest
COPY release/nhp-server/nhp-serverd /nhp-server/nhp-serverd
COPY release/nhp-server/plugins /nhp-server/plugins
DOCKER
docker build -t opennhp-ac:current -f - . <<'DOCKER'
FROM opennhp-ac:latest
COPY release/nhp-ac/nhp-acd /nhp-ac/nhp-acd
DOCKER
# relay has no plugins; a host cross-build overlay also works:
#   GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -C endpoints -o /tmp/relayd ./relay/main
#   docker build -t opennhp-relay:current ...  (COPY /tmp/relayd)
```

> Notes / gotchas:
> - Server plugins are Go plugins → require **CGO** and the **same build** as the
>   server binary. Cross-compiling on macOS gives `plugin: not implemented`;
>   build inside the linux container.
> - Bind-mounting a plugin's `etc/` dir **hides** the image's `config.toml`; the
>   mount under `multicluster/nhp-server-c2/plugins/example/etc/` therefore
>   includes both `config.toml` and `resource.toml`. Missing `config.toml`
>   breaks ASP registration (knock fails with error 52002).
> - Host 8080 may be taken by another project, so the relay is published on 8081.
