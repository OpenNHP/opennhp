# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0](https://github.com/OpenNHP/opennhp/compare/v0.6.0...v1.0.0) (2026-01-17)


### ⚠ BREAKING CHANGES

* **crypto:** The cipher scheme numeric values have changed. Existing configurations with DefaultCipherScheme=0 will now use CURVE instead of GMSM. To maintain GMSM, change the value to 1.

### Features

* add cool startup banners with version info for all NHP components ([c25d138](https://github.com/OpenNHP/opennhp/commit/c25d138e7e7da3c1b2fa5d2994b16b7a8e8c9089))
* add multi-language support, dual IP display, footer sponsor info ([086a437](https://github.com/OpenNHP/opennhp/commit/086a4371044285c99ea2592ec53ca6e6c410808f))
* add nhp-agent to Docker debugging environment ([a8fe76a](https://github.com/OpenNHP/opennhp/commit/a8fe76a70b9129913c67b7dac55a5359b2fbd97d))
* Added two new login options: QR code login and OTP Code authentication ([2170368](https://github.com/OpenNHP/opennhp/commit/21703687e6f8df14131203716ffded8c882db964))
* **cli:** add --json output flag for scriptability ([#1346](https://github.com/OpenNHP/opennhp/issues/1346)) ([634e3d8](https://github.com/OpenNHP/opennhp/commit/634e3d8c8b5406f06a87ff3fa2e07d88f4643937))


### Bug Fixes

* add bounds checking to prevent panics from malformed input ([#1358](https://github.com/OpenNHP/opennhp/issues/1358)) ([6508248](https://github.com/OpenNHP/opennhp/commit/65082485c68a9e851304312eae48f4587fafeea8))
* **ci:** Add default permissions to build-binaries workflow ([#1327](https://github.com/OpenNHP/opennhp/issues/1327)) ([8d90359](https://github.com/OpenNHP/opennhp/commit/8d90359d1b0a534ddf0c9f59b6f24834717966af))
* **ci:** Add explicit permissions to workflow ([#1320](https://github.com/OpenNHP/opennhp/issues/1320)) ([40901c8](https://github.com/OpenNHP/opennhp/commit/40901c8f35eb30d40055610ebd7f81a40cf6e08e))
* **ci:** remove duplicate permissions block in ubuntu-build workflow ([#1382](https://github.com/OpenNHP/opennhp/issues/1382)) ([61edb8e](https://github.com/OpenNHP/opennhp/commit/61edb8ee6dd09ae626bf6dfb24f8f727b3122509))
* **ci:** Skip latest-release job on pull requests ([2228f55](https://github.com/OpenNHP/opennhp/commit/2228f550790f4e7ce5f3bf89aca92a7b46085ed5))
* **ci:** Skip latest-release job on pull requests ([9855cd6](https://github.com/OpenNHP/opennhp/commit/9855cd61e8945b565b1c4c83ad0108ecf1fcdde3))
* **ci:** Update Go version and fix grpc-gcp-go dependency ([9e76211](https://github.com/OpenNHP/opennhp/commit/9e762115633a706d304251a7e0e44740df6b617b))
* **ci:** Update Go version to 1.24 in build-binaries workflow ([c152689](https://github.com/OpenNHP/opennhp/commit/c152689971b42f7c06767f772c617fce4de66380))
* **crypto:** Add proper error handling for crypto operations ([#1338](https://github.com/OpenNHP/opennhp/issues/1338)) ([4547009](https://github.com/OpenNHP/opennhp/commit/454700912a7737eee8284bf3b401258fdf5731ff))
* **crypto:** change default cipher scheme from GMSM to CURVE ([#1330](https://github.com/OpenNHP/opennhp/issues/1330)) ([96baabd](https://github.com/OpenNHP/opennhp/commit/96baabd75bbe98788d03e6fb170e05197a6bec81))
* demo templates ([1f50bf1](https://github.com/OpenNHP/opennhp/commit/1f50bf181f33a26863c8ddd9b47a0ee5adfed62b))
* **deps:** Upgrade grpc-gcp-go to v1.6.0 for grpc v1.78.0 compatibility ([ae58134](https://github.com/OpenNHP/opennhp/commit/ae58134e93964975e45ace768515146138ebc718))
* **docker:** Make GOPROXY configurable instead of hardcoded ([#1345](https://github.com/OpenNHP/opennhp/issues/1345)) ([8a472c4](https://github.com/OpenNHP/opennhp/commit/8a472c44e6b714521a83ad8ff2db1f0d9c4b217b)), closes [#1314](https://github.com/OpenNHP/opennhp/issues/1314)
* Handle Codecov report issue for the committed file endpoint/server/config.go ([b9e5885](https://github.com/OpenNHP/opennhp/commit/b9e5885a19f9c5133d7c2b4f392b82e6a794757a))
* image path issue in the doc ([4d9b4df](https://github.com/OpenNHP/opennhp/commit/4d9b4df6ab9beee81228d43fbdf1f35b6ac9a7dc))
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
* nhp-server update resource.toml to resolve bug ([2170368](https://github.com/OpenNHP/opennhp/commit/21703687e6f8df14131203716ffded8c882db964))
* Optimize quick start doc ([64cdee8](https://github.com/OpenNHP/opennhp/commit/64cdee8183f3c5be5d38680e5e5958d005dd037e))
* prevent panics discovered by fuzz testing ([6534cf6](https://github.com/OpenNHP/opennhp/commit/6534cf632eed48b546fc4a6b3e64230f2fdcd2c2))
* prevent panics in crypto, peer, and compression functions ([1b69b37](https://github.com/OpenNHP/opennhp/commit/1b69b375b83cd02acf76afb8b5a5f2db8579901b))
* rename remote.toml to remote.toml.example to disable etcd by default ([00b8bdb](https://github.com/OpenNHP/opennhp/commit/00b8bdb26ef9418721d3a3de7aea969065f8cb64))
* resolve build failure in "Build and Test Code on Ubuntu / build (pull_request)" workflow ([02da9ad](https://github.com/OpenNHP/opennhp/commit/02da9ad83c6fc4640bf18272ea87658c038ad750))
* **security:** Set Secure flag on session cookies ([#1319](https://github.com/OpenNHP/opennhp/issues/1319)) ([8892bf8](https://github.com/OpenNHP/opennhp/commit/8892bf8b0f7d95ee37741eeb5550221e43eb5a6d))
* **security:** Upgrade golang.org/x/crypto to v0.46.0 in examples/server_plugin ([7f3934a](https://github.com/OpenNHP/opennhp/commit/7f3934a4a4a69be07874b424bf512ec392c23a76))
* **security:** Upgrade golang.org/x/crypto to v0.46.0 in examples/server_plugin ([dfb1ee3](https://github.com/OpenNHP/opennhp/commit/dfb1ee30f6cb63fe6e396f7c98339af463d2f90b))
* sync plugin dependencies with endpoints to fix version mismatch ([4226842](https://github.com/OpenNHP/opennhp/commit/422684228cc3fdd582eec0018aa72e6e3b32db82))
* update GitHub Actions config to resolve build failures after Go 1.25 upgrade ([01b2fc1](https://github.com/OpenNHP/opennhp/commit/01b2fc1de9db9920c1f132b03e668db83e119b51))
* update GitHub Actions config to resolve build failures after Go 1.25 upgrade ([e5ce8be](https://github.com/OpenNHP/opennhp/commit/e5ce8be974b057498df1371448dc563c23ee5c2b))
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
