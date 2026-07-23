# CHANGELOG

<!-- version list -->

## v0.11.2 (2026-07-23)

### Bug Fixes

- Gate harness version metadata to v7
  ([#18](https://github.com/ditto-assistant/dittobench-datagen/pull/18),
  [`9c348ee`](https://github.com/ditto-assistant/dittobench-datagen/commit/9c348ee3b59edfc96324fb08f1c22c83fc248565))

- **release**: Automate immutable datagen releases
  ([#19](https://github.com/ditto-assistant/dittobench-datagen/pull/19),
  [`f8108f1`](https://github.com/ditto-assistant/dittobench-datagen/commit/f8108f17d7d1e2dabe8640447a3194eae8895039))


## v0.11.1 (2026-07-21)

### Documentation

- Remove em-dashes from bench-versions.md (clearer punctuation)
  ([#16](https://github.com/ditto-assistant/dittobench-datagen/pull/16),
  [`f5435d9`](https://github.com/ditto-assistant/dittobench-datagen/commit/f5435d978bcbdbb3e53f5a641b1c873b378eada7))

### Features

- **v6**: Soften declarative-ack — accept an acknowledgment, not only a value echo
  ([#17](https://github.com/ditto-assistant/dittobench-datagen/pull/17),
  [`55343ff`](https://github.com/ditto-assistant/dittobench-datagen/commit/55343ff961e36ebaa1c7db378dd7e5e0663ff5f3))


## v0.11.0 (2026-07-21)

### Features

- Bench_version 6 — stored-instruction injection (memory-as-data)
  ([#15](https://github.com/ditto-assistant/dittobench-datagen/pull/15),
  [`01b93d2`](https://github.com/ditto-assistant/dittobench-datagen/commit/01b93d2e788de9c5bf62b50f834d8e7a36dfcc5a))

- Bench_version 6 — stored-instruction injection (memory-as-data, plan 4.8)
  ([#15](https://github.com/ditto-assistant/dittobench-datagen/pull/15),
  [`01b93d2`](https://github.com/ditto-assistant/dittobench-datagen/commit/01b93d2e788de9c5bf62b50f834d8e7a36dfcc5a))

- V6 complexity cases 4.3/4.4/4.6 (non-verbatim, consolidation, multi-query)
  ([#15](https://github.com/ditto-assistant/dittobench-datagen/pull/15),
  [`01b93d2`](https://github.com/ditto-assistant/dittobench-datagen/commit/01b93d2e788de9c5bf62b50f834d8e7a36dfcc5a))


## v0.10.0 (2026-07-21)

### Bug Fixes

- 4 declarative chains at full — stabilize conversational-sanity small-N variance
  ([#13](https://github.com/ditto-assistant/dittobench-datagen/pull/13),
  [`154f561`](https://github.com/ditto-assistant/dittobench-datagen/commit/154f561cba54f116af27cd6cd65c1217fbfa36df))

### Documentation

- Explain bench_version 4 and the versioning policy
  ([#9](https://github.com/ditto-assistant/dittobench-datagen/pull/9),
  [`7ee7d46`](https://github.com/ditto-assistant/dittobench-datagen/commit/7ee7d4646beda9067fcff2e2e22afbd5ed20446b))

### Features

- Bench_version 5 Phase A — conversational-sanity gate, declarative writes, accept-set
  ([`f452d2a`](https://github.com/ditto-assistant/dittobench-datagen/commit/f452d2a06642ee1314a3abba5f5d3968428a2361))

- V5 anti-overfit rework + token-efficiency contract
  ([#13](https://github.com/ditto-assistant/dittobench-datagen/pull/13),
  [`154f561`](https://github.com/ditto-assistant/dittobench-datagen/commit/154f561cba54f116af27cd6cd65c1217fbfa36df))

- V5 anti-overfit rework + token-efficiency contract (follow-up to #12)
  ([#13](https://github.com/ditto-assistant/dittobench-datagen/pull/13),
  [`154f561`](https://github.com/ditto-assistant/dittobench-datagen/commit/154f561cba54f116af27cd6cd65c1217fbfa36df))


## v0.9.0 (2026-07-19)

### Code Style

- Gofmt protocol struct alignment
  ([#8](https://github.com/ditto-assistant/dittobench-datagen/pull/8),
  [`673cdef`](https://github.com/ditto-assistant/dittobench-datagen/commit/673cdeffc1ba4a4f8629aafd575ce710f2bab52a))

### Features

- Add bench_version 4, correcting v3 false positives
  ([#8](https://github.com/ditto-assistant/dittobench-datagen/pull/8),
  [`673cdef`](https://github.com/ditto-assistant/dittobench-datagen/commit/673cdeffc1ba4a4f8629aafd575ce710f2bab52a))


## v0.8.0 (2026-07-18)

### Code Style

- Gofmt the v2 snapshot's import block
  ([#2](https://github.com/ditto-assistant/dittobench-datagen/pull/2),
  [`7894f3b`](https://github.com/ditto-assistant/dittobench-datagen/commit/7894f3b4b33b0f880eb19f4292b0357d0542187c))

### Features

- Add immutable DittoBench v3 datasets
  ([#3](https://github.com/ditto-assistant/dittobench-datagen/pull/3),
  [`77049bc`](https://github.com/ditto-assistant/dittobench-datagen/commit/77049bcb2ab4bb469251c012e08801a0aebb7086))

- **datagen**: Anti-gaming hardening for DittoBench v3
  ([#2](https://github.com/ditto-assistant/dittobench-datagen/pull/2),
  [`7894f3b`](https://github.com/ditto-assistant/dittobench-datagen/commit/7894f3b4b33b0f880eb19f4292b0357d0542187c))

- **datagen**: Serve frozen v2 from a snapshot; v3 is the hardened release
  ([#2](https://github.com/ditto-assistant/dittobench-datagen/pull/2),
  [`7894f3b`](https://github.com/ditto-assistant/dittobench-datagen/commit/7894f3b4b33b0f880eb19f4292b0357d0542187c))


## v0.7.1 (2026-07-16)

### Documentation

- Don't name a private internal service in the public README
  ([`0001c24`](https://github.com/ditto-assistant/dittobench-datagen/commit/0001c24eac00726ca34a920ec11433388dd2311b))

- **readme**: Add cmd/gstudy to the Layout list
  ([`d26d60c`](https://github.com/ditto-assistant/dittobench-datagen/commit/d26d60c4120dca987a0c4917ab6b673c88b53291))

### Features

- Cmd/generate-service — HTTP wrapper + image build for platform dataset pinning
  ([#1](https://github.com/ditto-assistant/dittobench-datagen/pull/1),
  [`0172a8c`](https://github.com/ditto-assistant/dittobench-datagen/commit/0172a8c2b8e168e6eaf2a8870897e207b38e8bb2))


## v0.7.0 (2026-07-12)

### Features

- **datagen**: Emit several metamorphic families, select by run size
  ([`f868288`](https://github.com/ditto-assistant/dittobench-datagen/commit/f8682880ba343cb5d39b7f939762510d710a3da7))


## v0.6.0 (2026-07-11)

### Features

- **datagen**: Lifecycle chains, point-in-time modality, grader-audit tool; doc cleanup (v0.6.0)
  ([`28c6fbc`](https://github.com/ditto-assistant/dittobench-datagen/commit/28c6fbc1a79df1224fe644da92b794df28deeb96))


## v0.5.0 (2026-07-10)

### Bug Fixes

- **datagen**: Close no-model router leaks in item-A grammars
  ([`7c1676e`](https://github.com/ditto-assistant/dittobench-datagen/commit/7c1676e1e6ef8f4c714a39767e07f1fc66fb0afb))

### Documentation

- **protocol**: Note metamorphic_consistency is now folded into the composite
  ([`b6f934a`](https://github.com/ditto-assistant/dittobench-datagen/commit/b6f934a50a41a79c376323fc9848a7d8568ecdd0))

### Features

- Question-audit fixes and grader hardening (2026-07-10 audit)
  ([`62cb10d`](https://github.com/ditto-assistant/dittobench-datagen/commit/62cb10d6f91087b19b3c6fe503138f9c5196265b))

- **datagen**: CFG surface expander for the five audited low-variety categories
  ([`db4011f`](https://github.com/ditto-assistant/dittobench-datagen/commit/db4011f70cbe26d92a57399c39490367997ef75a))

- **datagen**: Metamorphic invariance twins become a j=3 sibling family (N2)
  ([`406dfb8`](https://github.com/ditto-assistant/dittobench-datagen/commit/406dfb8942b481a8b634192a6f2db7fb138dcf61))

### Testing

- **datagen**: Generalized router-keyword leak guard for grammar categories
  ([`3a3b152`](https://github.com/ditto-assistant/dittobench-datagen/commit/3a3b1527036b599a181fd6be498d23a9b43fbcc1))

- **datagen**: Pin automation_list grammar against router-keyword leaks
  ([`6775afa`](https://github.com/ditto-assistant/dittobench-datagen/commit/6775afa574cb6779811d473d6a844110eaf4ebc4))


## v0.4.0 (2026-07-10)

### Features

- **protocol**: Judge-free deterministic grading data
  ([`90cf9e3`](https://github.com/ditto-assistant/dittobench-datagen/commit/90cf9e3666b1c73c9e5a1d24d7b224104840ac86))


## v0.3.0 (2026-07-10)

### Features

- **protocol**: Judge-audit agreement counts in RunDetails
  ([`b97c08b`](https://github.com/ditto-assistant/dittobench-datagen/commit/b97c08b35e5f809f8301697e40afd15d130a2da8))


## v0.2.0 (2026-07-09)

### Features

- Per-version surface rotation (public treadmill)
  ([`5a7e826`](https://github.com/ditto-assistant/dittobench-datagen/commit/5a7e8268877b6ae7dfd232cfbb78cc9859760903))


## v0.1.0 (2026-07-09)

- Initial Release
