# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Data source `dreamhost_dns_record` for looking up specific DNS records
- Data source `dreamhost_dns_records` for listing and filtering DNS records
- DNS record validation for all supported record types
- Retry logic with exponential backoff for API operations
- Wait states for DNS record creation and deletion
- Cache invalidation on record modifications
- Comprehensive examples for all use cases
- Import functionality for existing DNS records

### Changed
- Improved error messages with specific field names
- Enhanced documentation with detailed usage examples
- Updated provider to use proper caching with invalidation

### Fixed
- Fixed documentation attribute naming inconsistency
- Fixed error messages showing wrong field names
- Added missing DNS record type validation
- Fixed version inconsistency in examples
- Fixed CNAME record trailing dot handling

### Security
- Removed hardcoded API keys from examples
- Removed unused docker-compose configuration with hardcoded passwords
- Updated placeholder tokens to be more obvious

## [0.0.1] - 2024-01-01

### Added
- Initial release
- Basic DNS record management (Create, Read, Delete)
- Support for A, AAAA, CNAME, MX, NS, PTR, TXT, SRV, NAPTR record types
- DreamHost API integration
- Basic caching implementation

[Unreleased]: https://github.com/aygp-dr/terraform-provider-dreamhost/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/aygp-dr/terraform-provider-dreamhost/releases/tag/v0.0.1