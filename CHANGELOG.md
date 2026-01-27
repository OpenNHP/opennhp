# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 (2026-01-27)


### ⚠ BREAKING CHANGES

* **crypto:** The cipher scheme numeric values have changed. Existing configurations with DefaultCipherScheme=0 will now use CURVE instead of GMSM. To maintain GMSM, change the value to 1.

### Features

* add cool startup banners with version info for all NHP components ([c25d138](https://github.com/OpenNHP/opennhp/commit/c25d138e7e7da3c1b2fa5d2994b16b7a8e8c9089))
* add Deploy demologin workflows ([053e3b5](https://github.com/OpenNHP/opennhp/commit/053e3b53a17993c9e88eec509c00946548cbe13d))
* add multi-language support, dual IP display, footer sponsor info ([086a437](https://github.com/OpenNHP/opennhp/commit/086a4371044285c99ea2592ec53ca6e6c410808f))
* add NHP-AC template deployment via jump host ([cf5e147](https://github.com/OpenNHP/opennhp/commit/cf5e147f503d3d95c720432d6d16ca4c2524129f))
* add nhp-acd binary deployment to NHP-AC servers ([d514575](https://github.com/OpenNHP/opennhp/commit/d514575329137ae4d08c8a350628238238c7c8d0))
* add nhp-acd deployment to NHP-AC servers ([5638ec8](https://github.com/OpenNHP/opennhp/commit/5638ec803a101e290d7eb44d6f61e6fec18cff55))
* add nhp-agent to Docker debugging environment ([a8fe76a](https://github.com/OpenNHP/opennhp/commit/a8fe76a70b9129913c67b7dac55a5359b2fbd97d))
* add platform selection checkboxes for build workflow ([fcb4cc6](https://github.com/OpenNHP/opennhp/commit/fcb4cc6e8b356e13c4f4a93775b05bc0b4b8acef))
* add platform selection for build workflow ([b33a11c](https://github.com/OpenNHP/opennhp/commit/b33a11c342b764f492695891fb1a2f430903cd43))
* add plugin build and deployment support to CI/CD workflows ([2478e8a](https://github.com/OpenNHP/opennhp/commit/2478e8a8c284f445e5c668db9c2b9232e0175726))
* add subdir support for external plugin repositories ([cac6f52](https://github.com/OpenNHP/opennhp/commit/cac6f52678d3380eea14743310f4fd380045f610))
* Add the OIDC authentication plugin. ([ee6ec8f](https://github.com/OpenNHP/opennhp/commit/ee6ec8f229c13d2375cbbc3ce0dcdfe5a5cbad00))
* Add the OIDC authentication plugin. ([f71277b](https://github.com/OpenNHP/opennhp/commit/f71277bf03e444ca7b6ab4a9a8a1eb4a1c89def4))
* Added two new login options: QR code login and OTP Code authentication ([2170368](https://github.com/OpenNHP/opennhp/commit/21703687e6f8df14131203716ffded8c882db964))
* auto-detect plugins in deploy-demo workflow ([979fa08](https://github.com/OpenNHP/opennhp/commit/979fa08948fe1499df8d2d2ce8e71b342a9a10d2))
* auto-discover and build all plugins in server_plugin directory ([cdabbb1](https://github.com/OpenNHP/opennhp/commit/cdabbb1f931b9b166765f7dc8aa33aa2dd4987aa))
* **cli:** add --json output flag for scriptability ([#1346](https://github.com/OpenNHP/opennhp/issues/1346)) ([634e3d8](https://github.com/OpenNHP/opennhp/commit/634e3d8c8b5406f06a87ff3fa2e07d88f4643937))
* Docker Local Debugging Environment ([2bde3e8](https://github.com/OpenNHP/opennhp/commit/2bde3e88479927568349ac5071dff463e2715656))
* Docker Local Debugging Environment ([#1267](https://github.com/OpenNHP/opennhp/issues/1267)) ([2bde3e8](https://github.com/OpenNHP/opennhp/commit/2bde3e88479927568349ac5071dff463e2715656))
* inject AUTH0 secrets from GitHub secrets in deploy workflow ([e17ea45](https://github.com/OpenNHP/opennhp/commit/e17ea45540b5bd34a976d12c1f84e232d179bcdb))
* make Build Binaries workflow manual-only with release options ([b205fce](https://github.com/OpenNHP/opennhp/commit/b205fce875f783a51ac5ce6858b82fde9c4de8c6))
* redesign basic plugin login page with SSO support ([9abcd93](https://github.com/OpenNHP/opennhp/commit/9abcd935b6ef91e3ad9c823a34e11661b8c31138))
* support nested plugin directories in Makefile ([b6cf9d0](https://github.com/OpenNHP/opennhp/commit/b6cf9d07855186a1b46cbe3ae7a90e2f38f6a257))


### Bug Fixes

* add bounds checking to prevent panics from malformed input ([#1358](https://github.com/OpenNHP/opennhp/issues/1358)) ([6508248](https://github.com/OpenNHP/opennhp/commit/65082485c68a9e851304312eae48f4587fafeea8))
* add checkout step and proper tag cleanup for latest release ([0b53a6d](https://github.com/OpenNHP/opennhp/commit/0b53a6d4c63c6649942fbdfb5f407d6f4d07fa83))
* add CUSTOM_LD_FLAGS parameter for make  ref [#1262](https://github.com/OpenNHP/opennhp/issues/1262) ([#1263](https://github.com/OpenNHP/opennhp/issues/1263)) ([d798c88](https://github.com/OpenNHP/opennhp/commit/d798c88a12ab8b750bf7977b162c6cb6915200f2))
* add go mod tidy before building plugins ([544fd84](https://github.com/OpenNHP/opennhp/commit/544fd84347b7fe1a03d5a621040b270927dbabc5))
* add ptype for port-scanner ([5a2d703](https://github.com/OpenNHP/opennhp/commit/5a2d703961c57add3167279fae3223088aebacec))
* address security vulnerabilities in oidc plugin ([f81e1f4](https://github.com/OpenNHP/opennhp/commit/f81e1f449562baa3e67280676ccfd04865e97bac))
* adjust go.mod replace paths for external plugins ([18e6ed8](https://github.com/OpenNHP/opennhp/commit/18e6ed83acc1644ebac082ac48d8db8dac06524f))
* **ci:** Add default permissions to build-binaries workflow ([#1327](https://github.com/OpenNHP/opennhp/issues/1327)) ([8d90359](https://github.com/OpenNHP/opennhp/commit/8d90359d1b0a534ddf0c9f59b6f24834717966af))
* **ci:** Add explicit permissions to workflow ([#1320](https://github.com/OpenNHP/opennhp/issues/1320)) ([40901c8](https://github.com/OpenNHP/opennhp/commit/40901c8f35eb30d40055610ebd7f81a40cf6e08e))
* cicd ([7b92747](https://github.com/OpenNHP/opennhp/commit/7b92747129ab05373410e442c03375a83fe974b3))
* **ci:** remove duplicate permissions block in ubuntu-build workflow ([#1382](https://github.com/OpenNHP/opennhp/issues/1382)) ([61edb8e](https://github.com/OpenNHP/opennhp/commit/61edb8ee6dd09ae626bf6dfb24f8f727b3122509))
* **ci:** Skip latest-release job on pull requests ([2228f55](https://github.com/OpenNHP/opennhp/commit/2228f550790f4e7ce5f3bf89aca92a7b46085ed5))
* **ci:** Skip latest-release job on pull requests ([9855cd6](https://github.com/OpenNHP/opennhp/commit/9855cd61e8945b565b1c4c83ad0108ecf1fcdde3))
* **ci:** Update Go version and fix grpc-gcp-go dependency ([9e76211](https://github.com/OpenNHP/opennhp/commit/9e762115633a706d304251a7e0e44740df6b617b))
* **ci:** Update Go version to 1.24 in build-binaries workflow ([c152689](https://github.com/OpenNHP/opennhp/commit/c152689971b42f7c06767f772c617fce4de66380))
* correct binary path after artifact download ([2504dc2](https://github.com/OpenNHP/opennhp/commit/2504dc2330067bd778a8cb2fa8b552889a02e0dc))
* correct external plugin clone path and support nested go.mod structure ([8ca46fc](https://github.com/OpenNHP/opennhp/commit/8ca46fcb6b29a0495e2ce3aad52aaff700197b74))
* correct oidc plugin output path in Makefile ([5354e62](https://github.com/OpenNHP/opennhp/commit/5354e62f113ef49ce2f1ef896dbd68289acdd9eb))
* correct service name from nhp-server to nhp-serverd in start command ([77c4553](https://github.com/OpenNHP/opennhp/commit/77c455311355e9cf48d923a883b9a69efa1f4711))
* correct template path for NHP-AC deployment ([b3f1009](https://github.com/OpenNHP/opennhp/commit/b3f1009691fe55e63399e8747957e61b7b09552a))
* **crypto:** Add proper error handling for crypto operations ([#1338](https://github.com/OpenNHP/opennhp/issues/1338)) ([4547009](https://github.com/OpenNHP/opennhp/commit/454700912a7737eee8284bf3b401258fdf5731ff))
* **crypto:** change default cipher scheme from GMSM to CURVE ([#1330](https://github.com/OpenNHP/opennhp/issues/1330)) ([96baabd](https://github.com/OpenNHP/opennhp/commit/96baabd75bbe98788d03e6fb170e05197a6bec81))
* demo templates ([1f50bf1](https://github.com/OpenNHP/opennhp/commit/1f50bf181f33a26863c8ddd9b47a0ee5adfed62b))
* deploy demo ([ca0542d](https://github.com/OpenNHP/opennhp/commit/ca0542d7a5d323a1c9a8b20dacec156c8ca17549))
* **deps:** Upgrade grpc-gcp-go to v1.6.0 for grpc v1.78.0 compatibility ([ae58134](https://github.com/OpenNHP/opennhp/commit/ae58134e93964975e45ace768515146138ebc718))
* **docker:** Make GOPROXY configurable instead of hardcoded ([#1345](https://github.com/OpenNHP/opennhp/issues/1345)) ([8a472c4](https://github.com/OpenNHP/opennhp/commit/8a472c44e6b714521a83ad8ff2db1f0d9c4b217b)), closes [#1314](https://github.com/OpenNHP/opennhp/issues/1314)
* fix escape mod bug ([#1257](https://github.com/OpenNHP/opennhp/issues/1257)) ([70166b4](https://github.com/OpenNHP/opennhp/commit/70166b46efba3725509a0e520acf89db3e156efc))
* Handle Codecov report issue for the committed file endpoint/server/config.go ([b9e5885](https://github.com/OpenNHP/opennhp/commit/b9e5885a19f9c5133d7c2b4f392b82e6a794757a))
* image path issue in the doc ([4d9b4df](https://github.com/OpenNHP/opennhp/commit/4d9b4df6ab9beee81228d43fbdf1f35b6ac9a7dc))
* improve latest release creation with gh CLI and retry handling ([4ec13b0](https://github.com/OpenNHP/opennhp/commit/4ec13b025970b9d98a8e71fb176aa4b826d916c2))
* improve OIDC security and login page UX ([20c64cc](https://github.com/OpenNHP/opennhp/commit/20c64cc55d0c3d1573ce2e6223e5d6c8cafa695b))
* **iptables:** fix rsyslog typo, permission issues and add log cleanup ([6e0f81c](https://github.com/OpenNHP/opennhp/commit/6e0f81c3f37a983501907ffea3696467782fd208))
* **ipv6:** fix critical bugs in IPv6 support implementation ([2261567](https://github.com/OpenNHP/opennhp/commit/22615674fc9fc7c46eb5bc0c14a37f0d125c3192))
* **ipv6:** fix critical bugs in IPv6 support implementation ([27e3d6b](https://github.com/OpenNHP/opennhp/commit/27e3d6bcfd2a2602635b46b83b678a485a362eaf))
* **ipv6:** implement full IPv6 support for iptables and AC ([#1329](https://github.com/OpenNHP/opennhp/issues/1329)) ([9ed0cf6](https://github.com/OpenNHP/opennhp/commit/9ed0cf6c059919e04de95c72c5747eff46c97022))
* **kgc:** secure master key file permissions ([#1337](https://github.com/OpenNHP/opennhp/issues/1337)) ([88b6f1f](https://github.com/OpenNHP/opennhp/commit/88b6f1fa0a953e636701d28a1b8dff830ee23226))
* **lint:** enable errcheck linter and fix all unchecked errors ([72c7e9a](https://github.com/OpenNHP/opennhp/commit/72c7e9ab41a419bc2829431a290048b11318c034))
* **lint:** enable errcheck linter and fix all unchecked errors ([2835f12](https://github.com/OpenNHP/opennhp/commit/2835f1237d9c68342b1a2eda3816004ed4064fc1))
* **lint:** fix additional errcheck issues and improve error handling ([4497d2f](https://github.com/OpenNHP/opennhp/commit/4497d2f6324c545ecbd7ca0a0d258da8724d607a))
* **lint:** fix loop variable capture bug in goroutine closures ([42f3940](https://github.com/OpenNHP/opennhp/commit/42f39409def7fc9ff016776784a55f204339f2ef))
* **log:** Add error handling for file sync and close operations ([#1318](https://github.com/OpenNHP/opennhp/issues/1318)) ([072594a](https://github.com/OpenNHP/opennhp/commit/072594a731a0a80997ab81621aea7325feafeae5))
* make build mandatory and restart service after templates update ([ad76813](https://github.com/OpenNHP/opennhp/commit/ad76813c86bb7b2da8d8646f1c7302ee40fe2ce5))
* nhp-server update resource.toml to resolve bug ([2170368](https://github.com/OpenNHP/opennhp/commit/21703687e6f8df14131203716ffded8c882db964))
* oidc plugins go.mod issue ([81d6e2b](https://github.com/OpenNHP/opennhp/commit/81d6e2bb59204932c07cf27bf29f5482caf0c83f))
* Optimize quick start doc ([64cdee8](https://github.com/OpenNHP/opennhp/commit/64cdee8183f3c5be5d38680e5e5958d005dd037e))
* plugin not supported on windows/amd64 ([baf845c](https://github.com/OpenNHP/opennhp/commit/baf845c17cc362321b8aa490c78486dc1edd5666))
* prevent panics discovered by fuzz testing ([6534cf6](https://github.com/OpenNHP/opennhp/commit/6534cf632eed48b546fc4a6b3e64230f2fdcd2c2))
* prevent panics in crypto, peer, and compression functions ([1b69b37](https://github.com/OpenNHP/opennhp/commit/1b69b375b83cd02acf76afb8b5a5f2db8579901b))
* rename remote.toml to remote.toml.example to disable etcd by default ([00b8bdb](https://github.com/OpenNHP/opennhp/commit/00b8bdb26ef9418721d3a3de7aea969065f8cb64))
* resolve build failure in "Build and Test Code on Ubuntu / build (pull_request)" workflow ([02da9ad](https://github.com/OpenNHP/opennhp/commit/02da9ad83c6fc4640bf18272ea87658c038ad750))
* rollback example_login.html ([f712f39](https://github.com/OpenNHP/opennhp/commit/f712f39f1f5b6453207bbc28d3911b9c3ff62e09))
* security vulnerabilities and bugs in CI/CD workflows ([edd595d](https://github.com/OpenNHP/opennhp/commit/edd595d07ae43ef1abe04c8985c7081b0dd9efa4))
* **security:** Set Secure flag on session cookies ([#1319](https://github.com/OpenNHP/opennhp/issues/1319)) ([8892bf8](https://github.com/OpenNHP/opennhp/commit/8892bf8b0f7d95ee37741eeb5550221e43eb5a6d))
* **security:** Upgrade golang.org/x/crypto to v0.46.0 in examples/server_plugin ([7f3934a](https://github.com/OpenNHP/opennhp/commit/7f3934a4a4a69be07874b424bf512ec392c23a76))
* **security:** Upgrade golang.org/x/crypto to v0.46.0 in examples/server_plugin ([dfb1ee3](https://github.com/OpenNHP/opennhp/commit/dfb1ee30f6cb63fe6e396f7c98339af463d2f90b))
* sync all oidc plugin dependencies with endpoints module ([9a28eed](https://github.com/OpenNHP/opennhp/commit/9a28eed1f0831a436ceb227b5dc511ed9f3fe4e6))
* sync oidc plugin Go version with nhp module ([e88c425](https://github.com/OpenNHP/opennhp/commit/e88c4252c62138366eacc3ef80888b9ff342917a))
* sync plugin dependencies with endpoints to fix version mismatch ([4226842](https://github.com/OpenNHP/opennhp/commit/422684228cc3fdd582eec0018aa72e6e3b32db82))
* templates ([c695383](https://github.com/OpenNHP/opennhp/commit/c69538328e007671ec4bb02a865b01b5a4d8a563))
* update GitHub Actions config to resolve build failures after Go 1.25 upgrade ([01b2fc1](https://github.com/OpenNHP/opennhp/commit/01b2fc1de9db9920c1f132b03e668db83e119b51))
* update GitHub Actions config to resolve build failures after Go 1.25 upgrade ([e5ce8be](https://github.com/OpenNHP/opennhp/commit/e5ce8be974b057498df1371448dc563c23ee5c2b))
* update oidc plugin go.mod to match other plugins ([ef1d868](https://github.com/OpenNHP/opennhp/commit/ef1d8688a0c58f5a3e6230fc678cc544d8e2cbb8))
* update oidc plugin oauth2 version to match endpoints ([3abf2e7](https://github.com/OpenNHP/opennhp/commit/3abf2e78dadc1faf6b0bc243521a4a48ddcfa452))
* update oidc plugin session dependencies to match endpoints ([316a9cd](https://github.com/OpenNHP/opennhp/commit/316a9cd736217c97d536854909eb594fbca73cff))
* update SSO and OTP login URLs ([8e79045](https://github.com/OpenNHP/opennhp/commit/8e79045b727bee455d4dd8948a7a396e4f450546))
* use direct sessions package in oidc plugin instead of reflection ([b7b65cc](https://github.com/OpenNHP/opennhp/commit/b7b65cc5050641dcb5b4d1338b4cd229ee6868ca))
* use reflection to access session in oidc plugin ([71efa20](https://github.com/OpenNHP/opennhp/commit/71efa206025d886b1acb81de7b2542adc406087b))
* use session helper callbacks to bypass Go plugin type limitation ([fd3a17e](https://github.com/OpenNHP/opennhp/commit/fd3a17e50c00b20285704d5054b3825b69fa187d))
* use URL-embedded token for private repo cloning ([7504f6f](https://github.com/OpenNHP/opennhp/commit/7504f6f99bd15b3aae39dc3be1e66784482cdde1))
* **utils:** use format specifiers in fmt.Errorf calls ([cfa028d](https://github.com/OpenNHP/opennhp/commit/cfa028db55392700292c53fe6e3409727c3309aa))
* **utils:** use format specifiers in fmt.Errorf calls ([c797179](https://github.com/OpenNHP/opennhp/commit/c7971799eb134b6fbcba2bb9232189456e10ce8d))

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
