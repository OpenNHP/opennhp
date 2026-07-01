# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0](https://github.com/OpenNHP/opennhp/compare/v1.0.0...v2.0.0) (2026-07-01)


### ⚠ BREAKING CHANGES

* **server,agent:** a server with this change rejects any agent that does not populate AgentKnockMsg.HeaderType. The agent half is included here; build and deploy agents from this change together with the server.
* **agent:** bind resources to clusters by name, drop ambiguous host fields
* **core,server:** stateless overload cookies for cluster verification
* **relay:** lift multi-instance restriction, pick by load balance scheme
* **core:** move recv-address stickiness from peer to ConnectionData

### Features

* **ac:** fan-out AOL/KPL to all server endpoints in a cluster ([77f1fa5](https://github.com/OpenNHP/opennhp/commit/77f1fa5a6628fcc9738866ce2a577027467d983d))
* **ac:** support multiple endpoints per nhp-server identity ([066bdf1](https://github.com/OpenNHP/opennhp/commit/066bdf11a22d73a4f0cce2e0f90f2457a7ca028f))
* **agent:** bind resources to clusters by name, drop ambiguous host fields ([a947e1c](https://github.com/OpenNHP/opennhp/commit/a947e1c5c3deedcdd2a16c5d7b47550f19e98708))
* **agent:** multi-instance server clusters with sticky/non-sticky modes ([187d1b3](https://github.com/OpenNHP/opennhp/commit/187d1b3c638d507b3e3b0f5c3c179a6c2a1ca598))
* **core,server:** stateless overload cookies for cluster verification ([d46bda1](https://github.com/OpenNHP/opennhp/commit/d46bda15f66e3f1f945b4bb9b7f4cfc0b3fa54c9))
* **core:** move recv-address stickiness from peer to ConnectionData ([9e40d88](https://github.com/OpenNHP/opennhp/commit/9e40d8808b5abe56e82e14e4080b2af43ca94342))
* **demo:** add demo.nhp TLS proxy to AC nginx ([cc0647c](https://github.com/OpenNHP/opennhp/commit/cc0647c9ab689524291d1fa1e80e6d45f07dc060))
* **demo:** add demo.nhp TLS proxy to AC nginx ([738a996](https://github.com/OpenNHP/opennhp/commit/738a996f061c8df9ee1e9bc62b8fc9781e451d27))
* **demo:** add second independent nhp-server+nhp-ac cluster (cluster 2) ([8cc75f7](https://github.com/OpenNHP/opennhp/commit/8cc75f715f0a9dfa11646d5f1edd072e7d17221e))
* **js-agent:** add CBOR token support for loading NHP-Agent parameters ([1fb9793](https://github.com/OpenNHP/opennhp/commit/1fb979382c69d2a289a4fa51b888e069da422cd9))
* **js-agent:** add CBOR token support for loading NHP-Agent parameters ([cc36a68](https://github.com/OpenNHP/opennhp/commit/cc36a684348b57bf348301bcf126f17d8a8c04e2))
* **relay:** lift multi-instance restriction, pick by load balance scheme ([66c1730](https://github.com/OpenNHP/opennhp/commit/66c1730c159f5dcc093ebb01097cac5295b0d203))
* **server:** add AllowPrivateRelaySource flag for local-demo NAT ([fe50276](https://github.com/OpenNHP/opennhp/commit/fe50276c35d84d8f2ceb6d0e31c6daa456c6aff3))
* **server:** add ForceOverload debug flag to exercise the cookie path ([43d1b6a](https://github.com/OpenNHP/opennhp/commit/43d1b6a8d421718d45fb06838edcf5e3eadf7466))
* **server:** per-source-IP rate limit for RKN-under-overload cookie ECDH ([2a3ef64](https://github.com/OpenNHP/opennhp/commit/2a3ef642e4451eb9add39d756ff06e83b922f381))


### Bug Fixes

* **ac,core:** post-review cleanup — lock config.Servers, skip empty SrcIP, comment fixes ([3e56ffc](https://github.com/OpenNHP/opennhp/commit/3e56ffc7879e98565d7cdaa99bd3d89439ee2630))
* **ac:** fail-close on initial etcd peer-table load ([16b7180](https://github.com/OpenNHP/opennhp/commit/16b7180988759ff56503a76a62e40e8fab70d872))
* address code review feedback for CBOR token modal ([c245470](https://github.com/OpenNHP/opennhp/commit/c245470358412b8650496aa58536d8ff4a916ade))
* **agent,ac,relay:** make config-reload paths preserve cluster bindings and fail-close ([f545851](https://github.com/OpenNHP/opennhp/commit/f545851bf555ee4c72a7defd8b44c7f6894d0fc1))
* **agent,ac,server:** close nil-deref, sticky-pin wedge, and attestation bypass ([e9be6c7](https://github.com/OpenNHP/opennhp/commit/e9be6c7a90742303f4532107fbe640bcae5a0a90))
* **agent,ac:** repair SDK knock crash and AC endpoint-key collision ([c6f9c76](https://github.com/OpenNHP/opennhp/commit/c6f9c76921affc738ef7eb65d61aa9fd151890c8))
* **agent:** clear stale cookie on cluster swap; non-blocking SDK signals ([568cb53](https://github.com/OpenNHP/opennhp/commit/568cb53fb9227a0bb62bbcee8707fc07517327b9))
* **agent:** guard SDK request sends against post-Stop channel close ([af93143](https://github.com/OpenNHP/opennhp/commit/af9314327fc04b05597d0014af0034be740fb02c))
* **agent:** prevent double-close and post-stop send panics in lifecycle ([8e983f1](https://github.com/OpenNHP/opennhp/commit/8e983f1d7be903c568d40349c987097a08c97243))
* **agent:** propagate specific cluster-lookup errors to SDK callers ([c318abd](https://github.com/OpenNHP/opennhp/commit/c318abd55b156ecbcc149525baea11fc57ac6b1f))
* **agent:** re-arm knock-stop Once per Start; guard ExitKnockRequest send ([bf3e5ef](https://github.com/OpenNHP/opennhp/commit/bf3e5efeb7e48880d6e39691685f529e406d738a))
* bump go-toml to v2.4.1 in plugin modules to match endpoints ([38459d9](https://github.com/OpenNHP/opennhp/commit/38459d999cabdf28c4b57c1f88dfdf420fd570ea))
* bump go-toml to v2.4.1 in plugin modules to match endpoints ([a3c3e8f](https://github.com/OpenNHP/opennhp/commit/a3c3e8f67e43b784c43c98b8cfab095493aa0c86))
* **core,server:** post-second-review correctness + comment hardening ([751529b](https://github.com/OpenNHP/opennhp/commit/751529bd4aaa320b8922360698cefee928720d45))
* **core:** copy cookie key out of StatelessCookieParams under the lock ([a4f7322](https://github.com/OpenNHP/opennhp/commit/a4f73223c3da93039bbd79a98dcc98dbed828c2a))
* **core:** key cookies on real client addr, not relay addr ([c828f75](https://github.com/OpenNHP/opennhp/commit/c828f75869192d98466e4fb7c6682996f2323209))
* **core:** remove cgo dependency from errors ([a7aa3e6](https://github.com/OpenNHP/opennhp/commit/a7aa3e63cbec97cefa3f0489ba48fa30ff2efffb))
* **core:** remove cgo dependency from errors ([6558da6](https://github.com/OpenNHP/opennhp/commit/6558da644bfe391ab9ac88c6375e15827c4b6f6d))
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
* **demo:** force fresh connection + no-store on acdemo vhosts ([1516585](https://github.com/OpenNHP/opennhp/commit/1516585879f4404bba7e6c440e4da57c8fabbac8))
* **demo:** grant nhp-acd CAP_NET_RAW and CAP_DAC_OVERRIDE ([3230f57](https://github.com/OpenNHP/opennhp/commit/3230f5791014d21bfb0c661fc388b0ee9edf62b7))
* **demo:** null-guard scanPorts labels in acdemo changeLanguage ([f7a5316](https://github.com/OpenNHP/opennhp/commit/f7a53168d2b3a110cd7b14c8cfef2a0a240b3081))
* **demo:** per-stack resource.toml overrides so shared default stays :443 ([1c08665](https://github.com/OpenNHP/opennhp/commit/1c086651ac44a6707955472f19ef5483f6eb6095))
* **demo:** scan cluster-2 server2/ac2 SSH host keys in infra apply ([b686494](https://github.com/OpenNHP/opennhp/commit/b6864942701f2b1b884ea1c50edbea2f94e602b1))
* **demo:** use relative paths in multicluster knock-test scripts ([29f93f2](https://github.com/OpenNHP/opennhp/commit/29f93f2e2f7d8eb8783d45a700d8e971c34fc6a9))
* **deploy:** rewrite ac/server.toml template to new [[Servers.Instances]] schema ([2f4f680](https://github.com/OpenNHP/opennhp/commit/2f4f680b1611a2a4b0f67ee21627b54a8c0177f1))
* **deps:** bump vite to ^8.0.16 in docs (Dependabot [#108](https://github.com/OpenNHP/opennhp/issues/108)) ([b8a898b](https://github.com/OpenNHP/opennhp/commit/b8a898b52e69578483ad2972106c134bcf689936))
* **deps:** bump vite to ^8.0.16 in docs to fix CVE server.fs.deny bypass ([5764feb](https://github.com/OpenNHP/opennhp/commit/5764feb1ddd617941c39f6405bb93eb38ffc7ba3))
* **deps:** fix basic-ftp and ws vulnerabilities in docs ([e5542bd](https://github.com/OpenNHP/opennhp/commit/e5542bdf3bf2279664689ec077c1014dc8bc2ffa))
* **endpoints/ac:** fail-close expandServerPeers when an entry's Endpoints all fail to parse ([488cc87](https://github.com/OpenNHP/opennhp/commit/488cc874c717942ef81876d9838a30ad17a2fc4e))
* **endpoints/server:** keep s.config.CookieSigningKeyBase64 in sync with the running device key on reload ([71124c7](https://github.com/OpenNHP/opennhp/commit/71124c75b84ed15320ae26c77fd95bf61c2ed957))
* **js-agent:** address code review feedback for CBOR token modal ([08598cc](https://github.com/OpenNHP/opennhp/commit/08598cc5b5915ba97137854003c44dcc225145ae))
* **js-agent:** authenticate knock HeaderType (mirror wire type in body) ([d10afec](https://github.com/OpenNHP/opennhp/commit/d10afec2612e68e2afaea9916b058ed65adb4022))
* **js-agent:** map protected-host picker to cluster by explicit index ([322fa9d](https://github.com/OpenNHP/opennhp/commit/322fa9d1f2056d2dbbcbc3e5c933bf46537cba93))
* **lint:** correct British-spelling misspells flagged by golangci-lint 2.7.2 ([c2e12d7](https://github.com/OpenNHP/opennhp/commit/c2e12d7b6b9446b6e5833011eb53b2bc4713f6e2))
* **lint:** satisfy golangci-lint 2.7.2 (gosec G404, misspell) ([cd2bb33](https://github.com/OpenNHP/opennhp/commit/cd2bb3333eb111333ee80d88a89583be2975c7e2))
* **lint:** use US spelling in comments flagged by misspell (analog, penalized) ([33cb8cf](https://github.com/OpenNHP/opennhp/commit/33cb8cf576d4bcb67c8e19a6d75ed8ee4e4009ab))
* **nhp/core:** route sendCookie through PrevParserData so the wire counter matches the agent KNK ([a4388e9](https://github.com/OpenNHP/opennhp/commit/a4388e9b4d1419f7853d0c4bf7f3d9f3a131afc3))
* **noise:** never carry zeroed intermediate chain key between packets ([e7886f8](https://github.com/OpenNHP/opennhp/commit/e7886f8ec675f7a06d3151c8da24900cdf5ced9f))
* **relay:** bound concurrent forwards per instance to cap pendingRequests ([48dd472](https://github.com/OpenNHP/opennhp/commit/48dd472b6ddceaa69a6a704b6df12b51afcac110))
* **relay:** scope instance-address dedupe to (pubkey, addr) ([1230386](https://github.com/OpenNHP/opennhp/commit/1230386fe643c72def19df9c1b65cb12117d7201))
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
* **terraform:** mark demo_nhp_cert output as sensitive ([e903f92](https://github.com/OpenNHP/opennhp/commit/e903f92ca05264621342547cdd71c126d4cde020))
* **terraform:** mark demo_nhp_cert output as sensitive ([3e4a817](https://github.com/OpenNHP/opennhp/commit/3e4a817269c445e9f92e2e1a216330c1aabf593b))
* **terraform:** mark derived-from-sensitive outputs as nonsensitive ([73dbc24](https://github.com/OpenNHP/opennhp/commit/73dbc24445f69484edd5b347a22cdd596c652564))
* **terraform:** mark derived-from-sensitive outputs as nonsensitive ([9501ba2](https://github.com/OpenNHP/opennhp/commit/9501ba234b182ba5f945c29b0b8bd33aaa894a91))


### Performance Improvements

* **server:** make rknRateLimiter eviction O(1) via random sampling ([f17fc19](https://github.com/OpenNHP/opennhp/commit/f17fc19ace024da4f5e42b64f70aabb90a2ff6ea))


### Reverts

* **relay:** drop the multi-instance rejection introduced by f545851 ([2b5b41f](https://github.com/OpenNHP/opennhp/commit/2b5b41fde35db63ea8c74f85074e6906eb293d6e))

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
