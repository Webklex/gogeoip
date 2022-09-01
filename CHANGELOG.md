# Changelog

All notable changes to `webklex/gogeoip` will be documented in this file.

Updates should follow the [Keep a CHANGELOG](http://keepachangelog.com/) principles.

## [UNRELEASED]
### Fixed
- NaN

### Added
- NaN


## [2.0.0] - 2022-09-01
This version is a complete overhaul of v1. I've rewritten everything. However, the old "/json/{optional host}" call is
still available and hasn't changed.

### Breaking changes
- The old config is not compatible with the new format.
- Most command argument names have changed


## [1.2.1] - 2020-01-21
### Fixed
- Rate limit interval option gets no longer ignored

### Added
- Setup wizard added


## [1.2.0] - 2020-01-20
- Country information extended
- Output structure reworked


## [1.1.0] - 2020-01-19
- ip2location db added
- Get ISP information
- Get domain information
- Get usage information
- Get proxy information
- Available config parameters extended


## [1.0.4] - 2020-01-18
- Population density omitted if empty or null
- Detect tor users


## [1.0.3] - 2020-01-18
- ASN information added


## [1.0.2] - 2020-01-18
- Optional user agent information added


## [1.0.1] - 2020-01-17
- Language parsing improved


## [1.0.0] - 2020-01-17
- Initial release
