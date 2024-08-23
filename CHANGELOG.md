<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry is required to include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The tag should consist of where the change is being made ex. (x/staking), (store)
The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"API Breaking" for breaking Protobuf, gRPC and REST routes or CLI commands.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Features
* (docs) [#326](https://github.com/sedaprotocol/seda-chain/pull/326) Swagger API documentation
* (x/tally) [#311](https://github.com/sedaprotocol/seda-chain/pull/311) New module `x/tally`
* (x/data-proxy) [#339](https://github.com/sedaprotocol/seda-chain/pull/339) New module `x/data-proxy`

### Bug Fixes
* (build) [#321](https://github.com/sedaprotocol/seda-chain/pull/321) Rosetta build tag (not to be used with static compilation) 

### API Breaking
* (x/wasm-storage) [#311](https://github.com/sedaprotocol/seda-chain/pull/311) "Overlay" renamed to "executor" and a few other API changes

### State Machine Breaking
* (x/tally) [#323](https://github.com/sedaprotocol/seda-chain/pull/323) Post data results in batches
