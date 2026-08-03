# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.7.2 (2026-08-03)


### ⚠ BREAKING CHANGES

* **server,agent:** a server with this change rejects any agent that does not populate AgentKnockMsg.HeaderType. The agent half is included here; build and deploy agents from this change together with the server.
* **agent:** bind resources to clusters by name, drop ambiguous host fields
* **core,server:** stateless overload cookies for cluster verification
* **relay:** lift multi-instance restriction, pick by load balance scheme
* **core:** move recv-address stickiness from peer to ConnectionData
* **relay:** remove legacy POST /relay; tighten routing contract

### Features

* **ac:** fan-out AOL/KPL to all server endpoints in a cluster ([77f1fa5](https://github.com/OpenNHP/opennhp/commit/77f1fa5a6628fcc9738866ce2a577027467d983d))
* **ac:** support multiple endpoints per nhp-server identity ([066bdf1](https://github.com/OpenNHP/opennhp/commit/066bdf11a22d73a4f0cce2e0f90f2457a7ca028f))
* agent registration with OTP/REG/RAK, SQLite keystore, SES SMTP ([7fc976e](https://github.com/OpenNHP/opennhp/commit/7fc976eb4caf9785188c9e0b6d610a42c43b8db7))
* **agent-register:** complete NHP agent registration flow via relay ([397812f](https://github.com/OpenNHP/opennhp/commit/397812f4d51ee74e229ddeeddee46a4bc8905c02))
* **agent:** add interactive CLI registration command (OTP → REG two-step flow) ([c5ec008](https://github.com/OpenNHP/opennhp/commit/c5ec0080253234f58414c945b38a9b1697a1e2cb))
* **agent:** bind resources to clusters by name, drop ambiguous host fields ([a947e1c](https://github.com/OpenNHP/opennhp/commit/a947e1c5c3deedcdd2a16c5d7b47550f19e98708))
* **agent:** match register CLI cipher-scheme labels to the reg page ([58fb601](https://github.com/OpenNHP/opennhp/commit/58fb60167d0d55938b9e736e25e8517c950a8b49))
* **agent:** multi-instance server clusters with sticky/non-sticky modes ([187d1b3](https://github.com/OpenNHP/opennhp/commit/187d1b3c638d507b3e3b0f5c3c179a6c2a1ca598))
* **agent:** show both derived public keys in register CLI regardless of scheme ([cd168aa](https://github.com/OpenNHP/opennhp/commit/cd168aa8bd7edccbd5fafe87e5c0b82f44dd6528))
* AWS demo deployment (infra, nginx+LE, js-agent, nhp-relay) ([0582e89](https://github.com/OpenNHP/opennhp/commit/0582e8983e91c7c8914ce7b40335d8f0afc8dbb5))
* **core,server:** stateless overload cookies for cluster verification ([d46bda1](https://github.com/OpenNHP/opennhp/commit/d46bda15f66e3f1f945b4bb9b7f4cfc0b3fa54c9))
* **core:** move recv-address stickiness from peer to ConnectionData ([9e40d88](https://github.com/OpenNHP/opennhp/commit/9e40d8808b5abe56e82e14e4080b2af43ca94342))
* **demo:** add demo.nhp TLS proxy to AC nginx ([cc0647c](https://github.com/OpenNHP/opennhp/commit/cc0647c9ab689524291d1fa1e80e6d45f07dc060))
* **demo:** add demo.nhp TLS proxy to AC nginx ([738a996](https://github.com/OpenNHP/opennhp/commit/738a996f061c8df9ee1e9bc62b8fc9781e451d27))
* **demo:** add second independent nhp-server+nhp-ac cluster (cluster 2) ([8cc75f7](https://github.com/OpenNHP/opennhp/commit/8cc75f715f0a9dfa11646d5f1edd072e7d17221e))
* **deploy:** support dual cipher schemes (Curve25519 + SM2) in demo ([fd34319](https://github.com/OpenNHP/opennhp/commit/fd34319a6ee8ee3ba28572b32aaeb31c17b9aee2))
* interactive agent registration CLI + dual-cipher (Curve25519 + SM2) demo support ([6acdb5d](https://github.com/OpenNHP/opennhp/commit/6acdb5d42481979e135ee5e29b593301e81907e8))
* **js-agent:** add CBOR token support for loading NHP-Agent parameters ([1fb9793](https://github.com/OpenNHP/opennhp/commit/1fb979382c69d2a289a4fa51b888e069da422cd9))
* **js-agent:** add CBOR token support for loading NHP-Agent parameters ([cc36a68](https://github.com/OpenNHP/opennhp/commit/cc36a684348b57bf348301bcf126f17d8a8c04e2))
* **js-agent:** add EN/zh-cn language switcher to demo page ([5a7f89d](https://github.com/OpenNHP/opennhp/commit/5a7f89dd9c28b6bf32d1839f9693fce521b6b610))
* **js-agent:** add EN/zh-cn language switcher to demo page ([ca4604e](https://github.com/OpenNHP/opennhp/commit/ca4604e3883b1d99e005f527b1d8638455c94979))
* **js-agent:** email as primary ID + CBOR token + 24h key expiry hint ([155cd1a](https://github.com/OpenNHP/opennhp/commit/155cd1a31f152cf69787a59a392092a93d47c629))
* **js-agent:** expose NHPAgent.version + bump VERSION to 0.7.3 ([e9ba4a4](https://github.com/OpenNHP/opennhp/commit/e9ba4a42a93d485bb22e95f1ca303c169ae51cad))
* **js-agent:** expose NHPAgent.version sourced from nhp/version/VERSION ([0616094](https://github.com/OpenNHP/opennhp/commit/061609405b84098ce8f9a158aedaee3db1feaec0))
* **js-agent:** improve demo page DX for SDK integrators ([0e3c561](https://github.com/OpenNHP/opennhp/commit/0e3c5617b50cefe4391ee71c9cfdc21bab9f551a))
* **js-agent:** improve demo page DX for SDK integrators ([2f718af](https://github.com/OpenNHP/opennhp/commit/2f718af28b8c1072311abd143e583298f65d44d0))
* **js-agent:** scope perf metric to network round trip; demo form polish ([1f5408d](https://github.com/OpenNHP/opennhp/commit/1f5408d752ba2a1a5037f3de0436d6f6760f34a3))
* **js-agent:** scope perf metric to network round trip; demo form polish ([8b4c869](https://github.com/OpenNHP/opennhp/commit/8b4c869ab227fdf86714f927351af2ad85671edd))
* **js-agent:** success overlay, IP footer, and code panel polish ([bd38ed7](https://github.com/OpenNHP/opennhp/commit/bd38ed7ae2675f170fe67f18afdaa261f4792c25))
* **js-agent:** success overlay, IP footer, and code panel polish ([b55f14c](https://github.com/OpenNHP/opennhp/commit/b55f14ced9bb89ec8061e40aa04703aba6094493))
* **js-agent:** vendor browser SDK into endpoints/js-agent ([6709d00](https://github.com/OpenNHP/opennhp/commit/6709d00cca413ababa246b93806ab31b2c4eaf6c))
* **plugin:** improve OTP verification email (code-in-subject, headers, injection-safe) ([4cc3f0b](https://github.com/OpenNHP/opennhp/commit/4cc3f0b67b77b89dc32fdfd18c5db25af35d2f30))
* **plugin:** include OTP code in email subject and add opennhp.org link ([c881e2b](https://github.com/OpenNHP/opennhp/commit/c881e2bccf5d833056bcf3b60e4c6c2f67e80ba3))
* **reg:** add NHP-OTP / NHP-REG protocol workflow diagram to reg page ([afee0f9](https://github.com/OpenNHP/opennhp/commit/afee0f9f40bbbbcd86a1f87595d94ee0f0d75353))
* **reg:** add page header "NHP-Agent Key Registration Workflow Demo" ([ed5c0aa](https://github.com/OpenNHP/opennhp/commit/ed5c0aa55611c9d5167b34b8aad0798243a0f23a))
* **registration:** add WebAuthn/FIDO2 hardware key support ([e114d55](https://github.com/OpenNHP/opennhp/commit/e114d55fba1aa5ec43dd8d2a4278d65600c6446b))
* **reg:** redesign registration page — 2-step flow, colorful protocol diagram, message spec ([2565d52](https://github.com/OpenNHP/opennhp/commit/2565d5208cfd5a94f015710b2219210c2f0a88cc))
* **reg:** replace message spec with tabbed API examples panel ([112ea9f](https://github.com/OpenNHP/opennhp/commit/112ea9f910e5ea06a341b3784918e7d0006e1aea))
* **reg:** replace message spec with tabbed API examples panel ([3e6345b](https://github.com/OpenNHP/opennhp/commit/3e6345b6614290821d9d54397c4fe16655c84b8e))
* **relay:** lift multi-instance restriction, pick by load balance scheme ([66c1730](https://github.com/OpenNHP/opennhp/commit/66c1730c159f5dcc093ebb01097cac5295b0d203))
* **relay:** support multiple nhp-server clusters via pubkey-derived id ([eef8b56](https://github.com/OpenNHP/opennhp/commit/eef8b56dedeb75e0302080b5a31c552df4d16f63))
* **relay:** support multiple nhp-server clusters via pubkey-derived id ([d083653](https://github.com/OpenNHP/opennhp/commit/d0836539febc6b9ea4b264d2b55d1a30d2749270))
* **server-plugin:** polish basic auth-plugin demo page ([2b7a3f9](https://github.com/OpenNHP/opennhp/commit/2b7a3f91936543d2d3466d45abf7a47d8267769f))
* **server-plugin:** polish basic auth-plugin demo page ([599eb8c](https://github.com/OpenNHP/opennhp/commit/599eb8cc9c4180da381c84f81eb7b0e9c06e1df0))
* **server:** add AllowPrivateRelaySource flag for local-demo NAT ([fe50276](https://github.com/OpenNHP/opennhp/commit/fe50276c35d84d8f2ceb6d0e31c6daa456c6aff3))
* **server:** add ForceOverload debug flag to exercise the cookie path ([43d1b6a](https://github.com/OpenNHP/opennhp/commit/43d1b6a8d421718d45fb06838edcf5e3eadf7466))
* **server:** per-source-IP rate limit for RKN-under-overload cookie ECDH ([2a3ef64](https://github.com/OpenNHP/opennhp/commit/2a3ef642e4451eb9add39d756ff06e83b922f381))


### Bug Fixes

* **ac,core:** post-review cleanup — lock config.Servers, skip empty SrcIP, comment fixes ([3e56ffc](https://github.com/OpenNHP/opennhp/commit/3e56ffc7879e98565d7cdaa99bd3d89439ee2630))
* **ac:** fail-close on initial etcd peer-table load ([16b7180](https://github.com/OpenNHP/opennhp/commit/16b7180988759ff56503a76a62e40e8fab70d872))
* add OTP generation cooldown and reorder identity check before OTP ([f3d583b](https://github.com/OpenNHP/opennhp/commit/f3d583be481b96d9ecfdbb64609e8dfad9f81855))
* add per-userId and distinct-device OTP rate limits to prevent cooldown bypass ([f4ea94a](https://github.com/OpenNHP/opennhp/commit/f4ea94a624a7965daf2b2f6cb33e96c8b5830e49))
* address code review feedback for CBOR token modal ([c245470](https://github.com/OpenNHP/opennhp/commit/c245470358412b8650496aa58536d8ff4a916ade))
* address code review findings for agent registration ([310bcab](https://github.com/OpenNHP/opennhp/commit/310bcabd22e85872c86ae24ee751422e17d51398))
* address review — enforce OTP-key binding, wire WebAuthn verification, mask SMTP secret ([b5dd9d3](https://github.com/OpenNHP/opennhp/commit/b5dd9d3260672aba2a7e55b80297ea185234af92))
* address review — WebAuthn fail-closed, verifier hardening, handshake key binding ([0520613](https://github.com/OpenNHP/opennhp/commit/0520613ca062a2fc561a69af05d0530112936475))
* **agent,ac,relay:** make config-reload paths preserve cluster bindings and fail-close ([f545851](https://github.com/OpenNHP/opennhp/commit/f545851bf555ee4c72a7defd8b44c7f6894d0fc1))
* **agent,ac,server:** close nil-deref, sticky-pin wedge, and attestation bypass ([e9be6c7](https://github.com/OpenNHP/opennhp/commit/e9be6c7a90742303f4532107fbe640bcae5a0a90))
* **agent,ac:** repair SDK knock crash and AC endpoint-key collision ([c6f9c76](https://github.com/OpenNHP/opennhp/commit/c6f9c76921affc738ef7eb65d61aa9fd151890c8))
* **agent:** clear stale cookie on cluster swap; non-blocking SDK signals ([568cb53](https://github.com/OpenNHP/opennhp/commit/568cb53fb9227a0bb62bbcee8707fc07517327b9))
* **agent:** guard SDK request sends against post-Stop channel close ([af93143](https://github.com/OpenNHP/opennhp/commit/af9314327fc04b05597d0014af0034be740fb02c))
* **agent:** preserve extension response errors ([71a36b4](https://github.com/OpenNHP/opennhp/commit/71a36b4bdef62d68e6986236411fab6c9fa01d37))
* **agent:** preserve extension response errors ([bdaad29](https://github.com/OpenNHP/opennhp/commit/bdaad29a8f7a45d62dc984ec11d66e3b8f22414c))
* **agent:** prevent double-close and post-stop send panics in lifecycle ([8e983f1](https://github.com/OpenNHP/opennhp/commit/8e983f1d7be903c568d40349c987097a08c97243))
* **agent:** propagate specific cluster-lookup errors to SDK callers ([c318abd](https://github.com/OpenNHP/opennhp/commit/c318abd55b156ecbcc149525baea11fc57ac6b1f))
* **agent:** re-arm knock-stop Once per Start; guard ExitKnockRequest send ([bf3e5ef](https://github.com/OpenNHP/opennhp/commit/bf3e5efeb7e48880d6e39691685f529e406d738a))
* **agent:** scope missing-config tolerance to register; make run fail loudly ([ae05338](https://github.com/OpenNHP/opennhp/commit/ae05338caee146ada0208d1d84375fa9e20975d4))
* **agent:** send email in NHP-OTP UserData from register CLI ([d3eaf6b](https://github.com/OpenNHP/opennhp/commit/d3eaf6b2870b641b0e715b8e6da2f21892243622))
* **bootstrap-tls:** add verbose SAN diagnostic logging ([972258c](https://github.com/OpenNHP/opennhp/commit/972258c3ff0a5f29074e8c3e125e65fd3bd74a0a))
* **bootstrap-tls:** force certbot to use correct lineage when expanding SANs ([e351be3](https://github.com/OpenNHP/opennhp/commit/e351be3b9c5cff3842eb6c6fe46b3e75de2e2af2))
* **bootstrap-tls:** include all existing SANs when expanding cert with --expand ([4010c70](https://github.com/OpenNHP/opennhp/commit/4010c70263ce6ef0dfbffaf4ee7d72bacef5418d))
* bounds-check cborSkipValue and recover panics in message handlers ([94a5ff6](https://github.com/OpenNHP/opennhp/commit/94a5ff6736f4102ee357b1b56d93807589c8ec98))
* bump go-toml to v2.4.1 in plugin modules to match endpoints ([38459d9](https://github.com/OpenNHP/opennhp/commit/38459d999cabdf28c4b57c1f88dfdf420fd570ea))
* bump go-toml to v2.4.1 in plugin modules to match endpoints ([a3c3e8f](https://github.com/OpenNHP/opennhp/commit/a3c3e8f67e43b784c43c98b8cfab095493aa0c86))
* **ci:** allow fork contributors to trigger claude-pr-review ([133d527](https://github.com/OpenNHP/opennhp/commit/133d5272bb92c4446969378a142decbaef24af43))
* **ci:** fix claude-pr-review for fork PRs with gh token scrub bypass ([0705438](https://github.com/OpenNHP/opennhp/commit/07054388904d6b56c89d76c667db8d8405087080))
* **ci:** hard-fail external-plugin step instead of silent no-op ([30b30fa](https://github.com/OpenNHP/opennhp/commit/30b30fa3809405f6867a166268dd891b4e935e68))
* **ci:** move CLAUDE_CODE_SUBPROCESS_ENV_SCRUB override to job level ([71a67a2](https://github.com/OpenNHP/opennhp/commit/71a67a288fe8d391a1ff898c31de446578a988a9))
* **core,server:** post-second-review correctness + comment hardening ([751529b](https://github.com/OpenNHP/opennhp/commit/751529bd4aaa320b8922360698cefee928720d45))
* **core:** copy cookie key out of StatelessCookieParams under the lock ([a4f7322](https://github.com/OpenNHP/opennhp/commit/a4f73223c3da93039bbd79a98dcc98dbed828c2a))
* **core:** key cookies on real client addr, not relay addr ([c828f75](https://github.com/OpenNHP/opennhp/commit/c828f75869192d98466e4fb7c6682996f2323209))
* **core:** remove cgo dependency from errors ([a7aa3e6](https://github.com/OpenNHP/opennhp/commit/a7aa3e63cbec97cefa3f0489ba48fa30ff2efffb))
* **core:** remove cgo dependency from errors ([6558da6](https://github.com/OpenNHP/opennhp/commit/6558da644bfe391ab9ac88c6375e15827c4b6f6d))
* correct SAN extraction in bootstrap-tls.sh ([635774f](https://github.com/OpenNHP/opennhp/commit/635774fc4a50daeddbaa04da87791388b3d2571b))
* default require_email_match to true in demo plugin ([4d083af](https://github.com/OpenNHP/opennhp/commit/4d083af885b64a590e87fed6de0fd18ad4ab2f66))
* **demo:** address code review feedback for demo.nhp TLS ([c2f23ac](https://github.com/OpenNHP/opennhp/commit/c2f23acde90f8d4be9b0b1778a410c6463cc6e66))
* **demo:** address follow-up PR review issues for demo.nhp TLS ([3a779a4](https://github.com/OpenNHP/opennhp/commit/3a779a42b33030b6e976fff5c4c8131574010556))
* **demo:** address follow-up PR review issues in demo.nhp renewal workflow ([6be6d4d](https://github.com/OpenNHP/opennhp/commit/6be6d4df38757c8a616b578c5803656cfb7e5c3e))
* **demo:** address follow-up review issues in TLS automation ([ae4b87f](https://github.com/OpenNHP/opennhp/commit/ae4b87f70782d02aebe855e035b7e7830febae51))
* **demo:** address PR review issues for demo.nhp TLS proxy ([d3cc4b5](https://github.com/OpenNHP/opennhp/commit/d3cc4b518e1bb355334927bb42f1c7d17c88999e))
* **demo:** address PR review issues for demo.nhp TLS proxy ([b599278](https://github.com/OpenNHP/opennhp/commit/b5992780cdf73dbcd88ea6892662a64cb44bc33b))
* **demo:** address PR review issues for optional demo.nhp TLS deployment ([5747bee](https://github.com/OpenNHP/opennhp/commit/5747bee0dfc914def84ea1b0f1b74c71a5ce35da))
* **demo:** address review follow-ups for demo.nhp TLS ([c8f6dbc](https://github.com/OpenNHP/opennhp/commit/c8f6dbc9a1432d99050b0159777c1708e59b774d))
* **demo:** convert install-demo-nhp-cert.sh to LF line endings ([14d1013](https://github.com/OpenNHP/opennhp/commit/14d101352185b4aea630878114873c29fcacc02d))
* **demo:** correct step ordering, full cert chain, and add -target comment ([f85ddd8](https://github.com/OpenNHP/opennhp/commit/f85ddd88d26c66133bbae8377e299c298273e1f6))
* **demo:** create logs (plural) dir before starting nhp daemons ([742530e](https://github.com/OpenNHP/opennhp/commit/742530e071c20d23a07ff51b52dd2621f0a312ba))
* **demo:** delete legacy certbot lineage instead of expanding it ([7673c29](https://github.com/OpenNHP/opennhp/commit/7673c29a5bb7e43f83ca505c31407d85c5aa4108))
* **demo:** force fresh connection + no-store on acdemo vhosts ([1516585](https://github.com/OpenNHP/opennhp/commit/1516585879f4404bba7e6c440e4da57c8fabbac8))
* **demo:** grant nhp-acd CAP_NET_RAW and CAP_DAC_OVERRIDE ([3230f57](https://github.com/OpenNHP/opennhp/commit/3230f5791014d21bfb0c661fc388b0ee9edf62b7))
* **demo:** migrate legacy certbot lineage when renaming primary domain ([526969e](https://github.com/OpenNHP/opennhp/commit/526969ed5f2740997cc5ace8353076bbc1c0e87d))
* **demo:** null-guard scanPorts labels in acdemo changeLanguage ([f7a5316](https://github.com/OpenNHP/opennhp/commit/f7a53168d2b3a110cd7b14c8cfef2a0a240b3081))
* **demo:** per-stack resource.toml overrides so shared default stays :443 ([1c08665](https://github.com/OpenNHP/opennhp/commit/1c086651ac44a6707955472f19ef5483f6eb6095))
* **demo:** scan cluster-2 server2/ac2 SSH host keys in infra apply ([b686494](https://github.com/OpenNHP/opennhp/commit/b6864942701f2b1b884ea1c50edbea2f94e602b1))
* **demo:** use relative paths in multicluster knock-test scripts ([29f93f2](https://github.com/OpenNHP/opennhp/commit/29f93f2e2f7d8eb8783d45a700d8e971c34fc6a9))
* **demo:** use sudo test for /etc/letsencrypt path checks ([3ab0391](https://github.com/OpenNHP/opennhp/commit/3ab0391c814eb39b465d02dc5a8428e6abd11ade))
* **deploy:** add envsubst for SMTP placeholders in server2 plugin config ([535e43e](https://github.com/OpenNHP/opennhp/commit/535e43ef330f12dd0c9e2cdf7b5fe5010de27d0c))
* **deploy:** align server2 AC_PRIVATE_IP with server1 pattern ([94fde22](https://github.com/OpenNHP/opennhp/commit/94fde2246d21dbffa08a0142084126f0cdcfdde4))
* **deploy:** fix SMTP config not rendered on server2 and add missing DB migration ([d8c0823](https://github.com/OpenNHP/opennhp/commit/d8c0823aea56b55d5fc38ba261c85db613002894))
* **deploy:** register SM2 public keys for js-agent in server agent.toml ([7950650](https://github.com/OpenNHP/opennhp/commit/795065026bece3ff00ae418f48865cc7df552c48))
* **deploy:** rewrite ac/server.toml template to new [[Servers.Instances]] schema ([2f4f680](https://github.com/OpenNHP/opennhp/commit/2f4f680b1611a2a4b0f67ee21627b54a8c0177f1))
* **deploy:** rewrite import map path in reg.html for production deployment ([b1c71d9](https://github.com/OpenNHP/opennhp/commit/b1c71d995828f5c3c0b0103ac6e055b52788ede5))
* **deps:** bump vite to ^8.0.16 in docs (Dependabot [#108](https://github.com/OpenNHP/opennhp/issues/108)) ([b8a898b](https://github.com/OpenNHP/opennhp/commit/b8a898b52e69578483ad2972106c134bcf689936))
* **deps:** bump vite to ^8.0.16 in docs to fix CVE server.fs.deny bypass ([5764feb](https://github.com/OpenNHP/opennhp/commit/5764feb1ddd617941c39f6405bb93eb38ffc7ba3))
* **deps:** fix basic-ftp and ws vulnerabilities in docs ([e5542bd](https://github.com/OpenNHP/opennhp/commit/e5542bdf3bf2279664689ec077c1014dc8bc2ffa))
* **endpoints/ac:** fail-close expandServerPeers when an entry's Endpoints all fail to parse ([488cc87](https://github.com/OpenNHP/opennhp/commit/488cc874c717942ef81876d9838a30ad17a2fc4e))
* **endpoints/server:** keep s.config.CookieSigningKeyBase64 in sync with the running device key on reload ([71124c7](https://github.com/OpenNHP/opennhp/commit/71124c75b84ed15320ae26c77fd95bf61c2ed957))
* gofmt formatting in nhpmsg.go ([01b9b38](https://github.com/OpenNHP/opennhp/commit/01b9b3815db3effdb669b3f3d3cb016819795506))
* guard register config overwrite; derive SM2 pubkey instead of rotating ([0f6ad4b](https://github.com/OpenNHP/opennhp/commit/0f6ad4b4cdaffff381c7876e01be4ff2d807af90))
* **infra:** drop nhp-acd from root to ec2-user + bounded capabilities ([8813ca7](https://github.com/OpenNHP/opennhp/commit/8813ca76512d3dc5b30ef18718680c8dad5ffd39))
* initialize EXPAND_FLAG before use in bootstrap-tls.sh ([e9c17d5](https://github.com/OpenNHP/opennhp/commit/e9c17d5e740d48b316ad7c3914c5336eca9b23fd))
* **js-agent:** address code review feedback for CBOR token modal ([08598cc](https://github.com/OpenNHP/opennhp/commit/08598cc5b5915ba97137854003c44dcc225145ae))
* **js-agent:** always load demo config from server, drop localStorage cache ([7f80de1](https://github.com/OpenNHP/opennhp/commit/7f80de1eddcdb819730008c9597f1a222645ee0b))
* **js-agent:** always load demo config from server, drop localStorage… ([f1eb543](https://github.com/OpenNHP/opennhp/commit/f1eb543573a7b65bd7f17372f994b5af6b8895c4))
* **js-agent:** authenticate knock HeaderType (mirror wire type in body) ([d10afec](https://github.com/OpenNHP/opennhp/commit/d10afec2612e68e2afaea9916b058ed65adb4022))
* **js-agent:** map protected-host picker to cluster by explicit index ([322fa9d](https://github.com/OpenNHP/opennhp/commit/322fa9d1f2056d2dbbcbc3e5c933bf46537cba93))
* **js-agent:** navigate reg Knock button to agent vhost via URL fragment ([981c6f5](https://github.com/OpenNHP/opennhp/commit/981c6f51840d30e3bcd3166b83124958824f6efb))
* **js-agent:** prevent protected server section from wrapping to two lines ([fbc4ae1](https://github.com/OpenNHP/opennhp/commit/fbc4ae18e37774cb45d97c3f66fc6f98abd041d2))
* **js-agent:** prevent protected server section from wrapping to two lines ([ca398f4](https://github.com/OpenNHP/opennhp/commit/ca398f48625e72230edafce557c2c392566be82e))
* **js-agent:** remove hardcoded relay URL default from reg.html ([ca6b540](https://github.com/OpenNHP/opennhp/commit/ca6b540998e5cbdf73b2693c63f7148208a5cc49))
* **js-agent:** shorten protected server text to prevent wrapping ([6a9d475](https://github.com/OpenNHP/opennhp/commit/6a9d4752939ff14f5cbc78a81a01244d09a6a615))
* **js-agent:** shorten protected server text to prevent wrapping ([d282fae](https://github.com/OpenNHP/opennhp/commit/d282fae22c036a1e47646802f7c2c71270bbe00a))
* **js-agent:** sync package-lock.json version with package.json ([811f1cf](https://github.com/OpenNHP/opennhp/commit/811f1cf47ba1125e2d403d8f41eb40296f63eb3e))
* **js-agent:** update i18n strings to match shortened text ([0daaba2](https://github.com/OpenNHP/opennhp/commit/0daaba2a4054ba601256bff6384ea24966868e14))
* **js-agent:** update i18n strings to match shortened text ([1000647](https://github.com/OpenNHP/opennhp/commit/1000647bb2759b9a5e93493ca8d7fd6e3656d9f1))
* **js-agent:** use Curve25519 fingerprint for relay routing in GMSM mode ([7fb23cb](https://github.com/OpenNHP/opennhp/commit/7fb23cb2cc081c48377f0b40b4ef8c16a65e4d78))
* **js-agent:** use RAK packet in registerPublicKey expiresAt tests ([4620032](https://github.com/OpenNHP/opennhp/commit/46200325bec5375c513a36b96b1a3d25778291ca))
* **keystore:** add incremental migration for otp_records.pub_key column ([f15ec9a](https://github.com/OpenNHP/opennhp/commit/f15ec9a29f6a8b17cfa0b0b396c4b585eacf2e02))
* **lint:** correct British-spelling misspells flagged by golangci-lint 2.7.2 ([c2e12d7](https://github.com/OpenNHP/opennhp/commit/c2e12d7b6b9446b6e5833011eb53b2bc4713f6e2))
* **lint:** satisfy golangci-lint 2.7.2 (gosec G404, misspell) ([cd2bb33](https://github.com/OpenNHP/opennhp/commit/cd2bb3333eb111333ee80d88a89583be2975c7e2))
* **lint:** use US spelling in comments flagged by misspell (analog, penalized) ([33cb8cf](https://github.com/OpenNHP/opennhp/commit/33cb8cf576d4bcb67c8e19a6d75ed8ee4e4009ab))
* **nginx:** serve /nhp-js/ bundle from parent dir on reg vhost ([6511851](https://github.com/OpenNHP/opennhp/commit/6511851fadc253ac87c63da6fb89a9ebad043baa))
* **nginx:** use try_files instead of alias to serve config.json on reg vhost ([4360699](https://github.com/OpenNHP/opennhp/commit/4360699107dd8aee73df9db14a501e9f0847714c))
* **nhp/core:** route sendCookie through PrevParserData so the wire counter matches the agent KNK ([a4388e9](https://github.com/OpenNHP/opennhp/commit/a4388e9b4d1419f7853d0c4bf7f3d9f3a131afc3))
* nil config dereference in relay routing test helper ([57ff354](https://github.com/OpenNHP/opennhp/commit/57ff354a272caa7ba881d7ef6ebde8823ae324a2))
* **noise:** never carry zeroed intermediate chain key between packets ([e7886f8](https://github.com/OpenNHP/opennhp/commit/e7886f8ec675f7a06d3151c8da24900cdf5ced9f))
* OTP rate-limit info leak and sweep retention=0 regression ([d580d9b](https://github.com/OpenNHP/opennhp/commit/d580d9b653d48e864247c577890725c5d391b0af))
* pass --expand to certbot when cert already exists ([a5fad5f](https://github.com/OpenNHP/opennhp/commit/a5fad5f18b585b5ea46de6e559301ea3c3b714c0))
* **plugins:** align shared deps with endpoints to fix plugin.Open ([a4acc5f](https://github.com/OpenNHP/opennhp/commit/a4acc5fca84d2bd577f236054ddbcf2950643e91))
* **plugins:** align shared deps with endpoints to fix plugin.Open ([5390ffa](https://github.com/OpenNHP/opennhp/commit/5390ffa15b49b884baa87167f162e828b64046f8))
* prevent DOM-based XSS in js-agent demo pages via innerHTML ([6f8e450](https://github.com/OpenNHP/opennhp/commit/6f8e4509b4dc005215263c561792439feca401bd))
* re-issue cert when EXTRA_DOMAINS missing from existing SAN ([fb1c4d6](https://github.com/OpenNHP/opennhp/commit/fb1c4d6cc9001bf80ab5054b08c639e95ec62081))
* **relay:** bound concurrent forwards per instance to cap pendingRequests ([48dd472](https://github.com/OpenNHP/opennhp/commit/48dd472b6ddceaa69a6a704b6df12b51afcac110))
* **relay:** bound naked channel sends to avoid handler-goroutine leaks ([372216c](https://github.com/OpenNHP/opennhp/commit/372216cd8db51702b3d53cb02d61af35cb9d662f))
* **relay:** scope instance-address dedupe to (pubkey, addr) ([1230386](https://github.com/OpenNHP/opennhp/commit/1230386fe643c72def19df9c1b65cb12117d7201))
* **relay:** spell "behavior" the American way to satisfy misspell linter ([768725a](https://github.com/OpenNHP/opennhp/commit/768725a132b9151c8ab5165c4940289f19f4d8b2))
* remove IAM user from ses.tf, use manually-provisioned SMTP creds ([cf39b4e](https://github.com/OpenNHP/opennhp/commit/cf39b4e3b37a0190c55a581c1fc296eb34448c0a))
* remove WebAuthn and restore OTP/REG flow ([5200e1a](https://github.com/OpenNHP/opennhp/commit/5200e1a3b8cbf2bb59195df06d255a1e70239f65))
* remove WebAuthn UI from reg.html ([4ce46f4](https://github.com/OpenNHP/opennhp/commit/4ce46f4d6d926132befecffb2e1b2e598541cf1c))
* repoint recv channel + restart routine on ReinitWithKey; flag plugin API break ([ba864ec](https://github.com/OpenNHP/opennhp/commit/ba864ec3ce490baa5885582bea37170cc6cda01f))
* resolve TTL test flake from Unix-second truncation race ([495c838](https://github.com/OpenNHP/opennhp/commit/495c8387a8b4552f43f0b37044cbccc6e5ba0f3c))
* **server,agent:** authenticate knock HeaderType (reject on-path header flips) ([cfa0871](https://github.com/OpenNHP/opennhp/commit/cfa08718c36e1549ac92ecb485ad5706652c4154))
* **server,agent:** authenticate knock HeaderType to block on-path header flips ([306a73f](https://github.com/OpenNHP/opennhp/commit/306a73f0ab14c9a6111ae988d1402acac33abb2c))
* **server:** bind stateless cookie to agent static pubkey ([8bb8244](https://github.com/OpenNHP/opennhp/commit/8bb824457fdeb7308a092f0a12ab1b0d6d6d5151))
* **server:** close races and unbounded growth on hot config + handler paths ([bc499d7](https://github.com/OpenNHP/opennhp/commit/bc499d7cfcdc5e8d934445c198cd8ab2c3699af9))
* **server:** identity-check map entry on conn teardown to match counter accounting ([5c23ba2](https://github.com/OpenNHP/opennhp/commit/5c23ba27de49f91e9365830ba694a0b0b6c9c57c))
* **server:** IPv6-safe relay conn key, prevent counter leak ([e89e566](https://github.com/OpenNHP/opennhp/commit/e89e566445415da1e49b02adb20c37007ff9ace5))
* **server:** make relay stale-replace and teardown counter-safe ([db4d376](https://github.com/OpenNHP/opennhp/commit/db4d376896bdcbada913cf04887fb36510037038))
* **server:** model HRF fresh-insert branch in stale-replace race tests ([1ba54e4](https://github.com/OpenNHP/opennhp/commit/1ba54e47c6836bbea4f93dc1704653efe8c10796))
* **server:** preserve random cookie key on window-only reload (single-instance) ([c0f68fd](https://github.com/OpenNHP/opennhp/commit/c0f68fd280ebb71da362ae10be8e6841c080d63b))
* **server:** raise guard-disabling flags to Critical and warn on demo-key hot-reload ([4e27e51](https://github.com/OpenNHP/opennhp/commit/4e27e5186d51404f2e323e22ce344093667742c1))
* **server:** reclaim per-relay slot on global-cap reject after stale-replace ([f5ef3fe](https://github.com/OpenNHP/opennhp/commit/f5ef3febc674baad21ae7601ca8341a566d17b1b))
* **server:** surface ForceOverload reload changes instead of silently dropping ([4a7eb89](https://github.com/OpenNHP/opennhp/commit/4a7eb89987596c6712843f3ad6d22b8ba66e8e6b))
* **server:** warn loudly when CookieSigningKeyBase64 matches the shipped demo value ([93553b4](https://github.com/OpenNHP/opennhp/commit/93553b4b3674816b465006ccf560dd179f391d85))
* smtp_port TOML validity and relay StickyInstance default ([871fc56](https://github.com/OpenNHP/opennhp/commit/871fc566d09472e39df216ed4b49c0a703bcf607))
* **terraform:** mark demo_nhp_cert output as sensitive ([e903f92](https://github.com/OpenNHP/opennhp/commit/e903f92ca05264621342547cdd71c126d4cde020))
* **terraform:** mark demo_nhp_cert output as sensitive ([3e4a817](https://github.com/OpenNHP/opennhp/commit/3e4a817269c445e9f92e2e1a216330c1aabf593b))
* **terraform:** mark derived-from-sensitive outputs as nonsensitive ([73dbc24](https://github.com/OpenNHP/opennhp/commit/73dbc24445f69484edd5b347a22cdd596c652564))
* **terraform:** mark derived-from-sensitive outputs as nonsensitive ([9501ba2](https://github.com/OpenNHP/opennhp/commit/9501ba234b182ba5f945c29b0b8bd33aaa894a91))
* use --force-renewal instead of --keep-until-expiring with --expand ([29911b0](https://github.com/OpenNHP/opennhp/commit/29911b096dd831e55868c3cb71389de1a9ee7a6d))
* use correct aws_ses_domain_identity attributes for provider v5 ([c6c3650](https://github.com/OpenNHP/opennhp/commit/c6c3650cabebd2a69410d8a2f9a2a5535479a989))
* use valid Go version in deploy workflow ([85b6fab](https://github.com/OpenNHP/opennhp/commit/85b6faba5d4c187df04585ee82063cd3c6d51232))


### Performance Improvements

* **server:** make rknRateLimiter eviction O(1) via random sampling ([f17fc19](https://github.com/OpenNHP/opennhp/commit/f17fc19ace024da4f5e42b64f70aabb90a2ff6ea))
* **server:** per-relay connection cap is now O(1) under a dedicated lock ([e7b74e6](https://github.com/OpenNHP/opennhp/commit/e7b74e634cb0d24523b98e976034c35fa9e6a94c))


### Reverts

* **relay:** drop the multi-instance rejection introduced by f545851 ([2b5b41f](https://github.com/OpenNHP/opennhp/commit/2b5b41fde35db63ea8c74f85074e6906eb293d6e))


### Miscellaneous Chores

* release 0.7.0 ([32957c2](https://github.com/OpenNHP/opennhp/commit/32957c245301e2c57ba9465b14ce2734c2158dde))
* release 0.7.1 ([15171e6](https://github.com/OpenNHP/opennhp/commit/15171e606c420fcbfb808936f773df4c037ed5a7))
* release 0.7.2 ([3e83d00](https://github.com/OpenNHP/opennhp/commit/3e83d00b596e3387cdfc75177362c6f825875fd2))


### Code Refactoring

* **relay:** remove legacy POST /relay; tighten routing contract ([bda88e1](https://github.com/OpenNHP/opennhp/commit/bda88e18108874bd38c02744f467cb37a4894321))

## [1.0.0](https://github.com/OpenNHP/opennhp/compare/v0.7.3...v1.0.0) (2026-05-15)


### ⚠ BREAKING CHANGES

* **relay:** remove legacy POST /relay; tighten routing contract

### Features

* **js-agent:** add EN/zh-cn language switcher to demo page ([5a7f89d](https://github.com/OpenNHP/opennhp/commit/5a7f89dd9c28b6bf32d1839f9693fce521b6b610))
* **js-agent:** add EN/zh-cn language switcher to demo page ([ca4604e](https://github.com/OpenNHP/opennhp/commit/ca4604e3883b1d99e005f527b1d8638455c94979))
* **js-agent:** expose NHPAgent.version + bump VERSION to 0.7.3 ([e9ba4a4](https://github.com/OpenNHP/opennhp/commit/e9ba4a42a93d485bb22e95f1ca303c169ae51cad))
* **js-agent:** expose NHPAgent.version sourced from nhp/version/VERSION ([0616094](https://github.com/OpenNHP/opennhp/commit/061609405b84098ce8f9a158aedaee3db1feaec0))
* **js-agent:** scope perf metric to network round trip; demo form polish ([1f5408d](https://github.com/OpenNHP/opennhp/commit/1f5408d752ba2a1a5037f3de0436d6f6760f34a3))
* **js-agent:** scope perf metric to network round trip; demo form polish ([8b4c869](https://github.com/OpenNHP/opennhp/commit/8b4c869ab227fdf86714f927351af2ad85671edd))
* **relay:** support multiple nhp-server clusters via pubkey-derived id ([eef8b56](https://github.com/OpenNHP/opennhp/commit/eef8b56dedeb75e0302080b5a31c552df4d16f63))
* **relay:** support multiple nhp-server clusters via pubkey-derived id ([d083653](https://github.com/OpenNHP/opennhp/commit/d0836539febc6b9ea4b264d2b55d1a30d2749270))


### Bug Fixes

* **js-agent:** prevent protected server section from wrapping to two lines ([fbc4ae1](https://github.com/OpenNHP/opennhp/commit/fbc4ae18e37774cb45d97c3f66fc6f98abd041d2))
* **js-agent:** prevent protected server section from wrapping to two lines ([ca398f4](https://github.com/OpenNHP/opennhp/commit/ca398f48625e72230edafce557c2c392566be82e))
* **js-agent:** shorten protected server text to prevent wrapping ([6a9d475](https://github.com/OpenNHP/opennhp/commit/6a9d4752939ff14f5cbc78a81a01244d09a6a615))
* **js-agent:** shorten protected server text to prevent wrapping ([d282fae](https://github.com/OpenNHP/opennhp/commit/d282fae22c036a1e47646802f7c2c71270bbe00a))
* **js-agent:** sync package-lock.json version with package.json ([811f1cf](https://github.com/OpenNHP/opennhp/commit/811f1cf47ba1125e2d403d8f41eb40296f63eb3e))
* **js-agent:** update i18n strings to match shortened text ([0daaba2](https://github.com/OpenNHP/opennhp/commit/0daaba2a4054ba601256bff6384ea24966868e14))
* **js-agent:** update i18n strings to match shortened text ([1000647](https://github.com/OpenNHP/opennhp/commit/1000647bb2759b9a5e93493ca8d7fd6e3656d9f1))
* **relay:** spell "behavior" the American way to satisfy misspell linter ([768725a](https://github.com/OpenNHP/opennhp/commit/768725a132b9151c8ab5165c4940289f19f4d8b2))


### Code Refactoring

* **relay:** remove legacy POST /relay; tighten routing contract ([bda88e1](https://github.com/OpenNHP/opennhp/commit/bda88e18108874bd38c02744f467cb37a4894321))

## [0.7.3](https://github.com/OpenNHP/opennhp/compare/v0.7.2...v0.7.3) (2026-05-08)


### Bug Fixes

* **js-agent:** always load demo config from server, drop localStorage cache ([7f80de1](https://github.com/OpenNHP/opennhp/commit/7f80de1eddcdb819730008c9597f1a222645ee0b))
* **js-agent:** always load demo config from server, drop localStorage… ([f1eb543](https://github.com/OpenNHP/opennhp/commit/f1eb543573a7b65bd7f17372f994b5af6b8895c4))
* **plugins:** align shared deps with endpoints to fix plugin.Open ([a4acc5f](https://github.com/OpenNHP/opennhp/commit/a4acc5fca84d2bd577f236054ddbcf2950643e91))
* **plugins:** align shared deps with endpoints to fix plugin.Open ([5390ffa](https://github.com/OpenNHP/opennhp/commit/5390ffa15b49b884baa87167f162e828b64046f8))

## [0.7.2](https://github.com/OpenNHP/opennhp/compare/v0.7.1...v0.7.2) (2026-05-07)


### Features

* **js-agent:** success overlay, IP footer, and code panel polish ([bd38ed7](https://github.com/OpenNHP/opennhp/commit/bd38ed7ae2675f170fe67f18afdaa261f4792c25))
* **js-agent:** success overlay, IP footer, and code panel polish ([b55f14c](https://github.com/OpenNHP/opennhp/commit/b55f14ced9bb89ec8061e40aa04703aba6094493))
* **server-plugin:** polish basic auth-plugin demo page ([2b7a3f9](https://github.com/OpenNHP/opennhp/commit/2b7a3f91936543d2d3466d45abf7a47d8267769f))
* **server-plugin:** polish basic auth-plugin demo page ([599eb8c](https://github.com/OpenNHP/opennhp/commit/599eb8cc9c4180da381c84f81eb7b0e9c06e1df0))


### Miscellaneous Chores

* release 0.7.2 ([3e83d00](https://github.com/OpenNHP/opennhp/commit/3e83d00b596e3387cdfc75177362c6f825875fd2))

## [Unreleased]

## [0.6.0] - 2025-06-11

### Added
- eBPF/XDP packet filtering support for high-performance knocking
- Docker local debugging environment
- `PASS_KNOCKIP_WITH_RANGE` mode for AC to include IP address ranges

### Changed
- Refactored peer hostname resolve logic
- Aligned UDP open resource behavior with HTTP version
- Server now continues when AC connections are lost in resource groups

### Fixed
- CGO compilation issues
- Escape mod bug
- Possible nil pointer dereference
- Size comparison error

## [0.5.0] - 2025-04-13

### Added
- Plugin system for NHP-Server with separate modules
- Improved build system for server plugins

### Changed
- Separated modules to accommodate building of nhp-serverd and its plugins

## [0.4.1] - 2025-04-06

### Added
- DHP (Data Hiding Protocol) function code
- SM2 P256 ECDH curve support
- Default cipher scheme configuration for DE

### Changed
- Using GMSM as default cipher scheme
- Updated Makefile for building DE on Linux

### Fixed
- Removed redundant logging
- Fixed SM2 P256 ECDH curve usage

## [0.4.0] - 2024-09-04

### Added
- Initial public release
- Jekyll-based documentation site
- GitHub Pages deployment

### Changed
- Updated code structure and symbols to be more self-explanatory

## [0.3.6] - 2024-09-03

### Added
- Pre-release version with core NHP protocol implementation
- Agent, Server, and AC components
- Noise Protocol Framework integration
- Curve25519 and SM2 cipher scheme support

[Unreleased]: https://github.com/OpenNHP/opennhp/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/OpenNHP/opennhp/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/OpenNHP/opennhp/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/OpenNHP/opennhp/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/OpenNHP/opennhp/compare/v0.3.6...v0.4.0
[0.3.6]: https://github.com/OpenNHP/opennhp/releases/tag/v0.3.6
