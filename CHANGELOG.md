# Changelog

All notable changes to PR Guardian are documented here.

## [1.0.0] — 2026-07-11

### Added
- Initial release of PR Guardian GitHub Action
- 7 automated PR checks: README, issue link, changelog, contributor guide, PR description, branch name, broken links
- AWS Lambda backend for the live web checker
- Web dashboard (landing page + live PR checker) at `web/`
- Docker image for GitHub Action container
- Unit tests for all locally-testable checks
