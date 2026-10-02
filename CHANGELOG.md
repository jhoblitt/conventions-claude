## [1.6.0](https://github.com/jhoblitt/conventions-claude/compare/v1.5.1...v1.6.0) (2026-10-02)

### Features

* **code-conventions:** link code citations in durable documents to a pinned line range ([883e61a](https://github.com/jhoblitt/conventions-claude/commit/883e61a9a823c6cf389aea53ab9e480e7e9fddb8)), closes [#23](https://github.com/jhoblitt/conventions-claude/issues/23)
* **github-conventions:** keep closing keywords away from references not meant to close ([e65c7ba](https://github.com/jhoblitt/conventions-claude/commit/e65c7ba08c45a4ec8724f635803eeacbb053a911)), closes [owner/repo#N](https://github.com/owner/repo/issues/N) [#N](https://github.com/jhoblitt/conventions-claude/issues/N) [#25](https://github.com/jhoblitt/conventions-claude/issues/25)
* **github-conventions:** link every item a PR description names ([747a237](https://github.com/jhoblitt/conventions-claude/commit/747a2379df434fa602fb070c190e59acf30279c1)), closes [#N](https://github.com/jhoblitt/conventions-claude/issues/N) [#N](https://github.com/jhoblitt/conventions-claude/issues/N) [#24](https://github.com/jhoblitt/conventions-claude/issues/24)

### Documentation

* record the linux-only release targets and the image opt-in ([7f122b3](https://github.com/jhoblitt/conventions-claude/commit/7f122b348939a681b7678ede3e6aba3a31b676d0)), closes [#13](https://github.com/jhoblitt/conventions-claude/issues/13)


## What's Changed
* chore(deps): bump the go-dependencies group across 2 directories with 1 update by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/28
* chore(deps): bump the github-actions group with 3 updates by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/29
* Prose rules: closing keywords, linked references, code citations, release record by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/26

## [1.5.1](https://github.com/jhoblitt/conventions-claude/compare/v1.5.0...v1.5.1) (2026-10-01)

### Bug Fixes

* **go-conventions:** derive the release fork guard from origin ([05f2e94](https://github.com/jhoblitt/conventions-claude/commit/05f2e94bb63fa11fc1f8f1339059d477fb5f538f)), closes [#14](https://github.com/jhoblitt/conventions-claude/issues/14)


## What's Changed
* fix(go-conventions): derive the release fork guard from origin by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/27

### Resolved issues

* [#14](https://github.com/jhoblitt/conventions-claude/issues/14) go-conventions: the release fork guard is derived from the module path, not the remote

## [1.5.0](https://github.com/jhoblitt/conventions-claude/compare/v1.4.0...v1.5.0) (2026-09-29)

### Features

* **code-conventions:** search for an existing report before reproducing ([36bcb6e](https://github.com/jhoblitt/conventions-claude/commit/36bcb6eda52f93bcf0d9e11136b7d16c30f10f6e))
* **code-conventions:** send a vulnerability to the project's private channel ([a32e468](https://github.com/jhoblitt/conventions-claude/commit/a32e4688458332a7644d4797a0d2d529085b16a9))
* **code-conventions:** write a reproducer in the upstream project's language ([4f3b430](https://github.com/jhoblitt/conventions-claude/commit/4f3b43011c19df0565f0f0a163e314bbdb6decd7))
* **github-conventions:** filing an issue waits for an instruction ([293adec](https://github.com/jhoblitt/conventions-claude/commit/293adecf0cd00361b7627300974de12c255a457b))
* **github-conventions:** rank the harness's defaults below every rung ([c2a4cf7](https://github.com/jhoblitt/conventions-claude/commit/c2a4cf750a4f1139f536f6f6af427e71cd81e709))

### Documentation

* state each procedural skill's approval gate in the README ([7955b14](https://github.com/jhoblitt/conventions-claude/commit/7955b14a9e7dbd124fa6c5e13a105c421e54255e))


## What's Changed
* Gate issue filing, private vulnerability reports, prior-art search, upstream-language reproducers by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/22

## [1.4.0](https://github.com/jhoblitt/conventions-claude/compare/v1.3.0...v1.4.0) (2026-09-29)

### Features

* **code-conventions:** bug reports carry a reproducer run before filing ([bdf43c8](https://github.com/jhoblitt/conventions-claude/commit/bdf43c8e1a3373081161533cfc51519e7ec8e20d)), closes [#20](https://github.com/jhoblitt/conventions-claude/issues/20)
* **github-conventions:** keep session links out of commits and PRs, the footer out of PRs ([6dbe865](https://github.com/jhoblitt/conventions-claude/commit/6dbe8653926b55856e46240350587811d360b31e)), closes [jhoblitt/go-ceph#14](https://github.com/jhoblitt/go-ceph/issues/14) [#18](https://github.com/jhoblitt/conventions-claude/issues/18) [#19](https://github.com/jhoblitt/conventions-claude/issues/19)


## What's Changed
* chore(deps): bump the github-actions group with 3 updates by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/16
* chore(deps): bump the go-dependencies group across 2 directories with 2 updates by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/17
* Keep session links and the Claude Code footer out; bug reports carry a reproducer by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/21

### Resolved issues

* [#18](https://github.com/jhoblitt/conventions-claude/issues/18) github-conventions: PR descriptions carry no claude.ai session link
* [#19](https://github.com/jhoblitt/conventions-claude/issues/19) github-conventions: PR descriptions carry no "Generated with Claude Code" footer
* [#20](https://github.com/jhoblitt/conventions-claude/issues/20) code-conventions: bug reports carry no reproduction instructions

## [1.3.0](https://github.com/jhoblitt/conventions-claude/compare/v1.2.2...v1.3.0) (2026-09-11)

### Features

* **code-conventions:** comments and docs describe the present ([38b9928](https://github.com/jhoblitt/conventions-claude/commit/38b992854000e56292f89c5dc71d012d8bd3b680)), closes [#9](https://github.com/jhoblitt/conventions-claude/issues/9)
* **github-conventions:** delete a merged PR's branch on new repositories ([491aeba](https://github.com/jhoblitt/conventions-claude/commit/491aeba93056ed7fad05de7eeba79f43e0c60aca)), closes [#8](https://github.com/jhoblitt/conventions-claude/issues/8)
* **go-conventions:** build linux only and make the image opt-in ([77580a0](https://github.com/jhoblitt/conventions-claude/commit/77580a073129ee6b9387359cc0266cd0870130b4)), closes [#5](https://github.com/jhoblitt/conventions-claude/issues/5)

### Bug Fixes

* **github-conventions:** let commitlint accept sentence-case subjects ([cf47bb7](https://github.com/jhoblitt/conventions-claude/commit/cf47bb78d95f5a6feea09639c957342bcbe4cdc9)), closes [#6](https://github.com/jhoblitt/conventions-claude/issues/6)


## What's Changed
* Resolve the four open issues: doc tense, Dependabot commitlint, delete-branch-on-merge, linux-only releases with image opt-in by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/12

### Resolved issues

* [#5](https://github.com/jhoblitt/conventions-claude/issues/5) go-conventions: goreleaser template should not build darwin and windows by default
* [#6](https://github.com/jhoblitt/conventions-claude/issues/6) github-conventions: commitlint template rejects Dependabot's sentence-case subjects
* [#8](https://github.com/jhoblitt/conventions-claude/issues/8) github-conventions: new repositories are not set to delete branches on merge
* [#9](https://github.com/jhoblitt/conventions-claude/issues/9) code-conventions: comments and docs describe the present, not the change

## [1.2.2](https://github.com/jhoblitt/conventions-claude/compare/v1.2.1...v1.2.2) (2026-09-11)

### Bug Fixes

* **github-conventions:** bump upload-artifact to v7 in the scorecard template ([2e379b1](https://github.com/jhoblitt/conventions-claude/commit/2e379b13470aa8d8d7f7bf393601fab3f25c70f6))
* **go-conventions:** bump upload-artifact to v7 in the ci template ([dbca972](https://github.com/jhoblitt/conventions-claude/commit/dbca97264979928a947408778fbe39f9415bdc1a))


## What's Changed
* chore(deps): bump the go-dependencies group across 2 directories with 2 updates by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/7
* chore(deps): bump actions/upload-artifact from 4.6.2 to 7.0.1 in the github-actions group across 1 directory by @dependabot[bot] in https://github.com/jhoblitt/conventions-claude/pull/3

## New Contributors
* @dependabot[bot] made their first contribution in https://github.com/jhoblitt/conventions-claude/pull/7

## [1.2.1](https://github.com/jhoblitt/conventions-claude/compare/v1.2.0...v1.2.1) (2026-09-11)

### Documentation

* **marketplace:** name both go-conventions dependencies ([fee4c73](https://github.com/jhoblitt/conventions-claude/commit/fee4c73217eb747398d4827854c51e255ffedc2e))


## What's Changed
* docs(marketplace): name both go-conventions dependencies by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/11

## [1.2.0](https://github.com/jhoblitt/conventions-claude/compare/v1.1.0...v1.2.0) (2026-09-11)

### Features

* **code-conventions:** add the chat-links reference ([7385c80](https://github.com/jhoblitt/conventions-claude/commit/7385c8036de74a8f3a43836fff4f7c2f1d948547))


## What's Changed
* feat(code-conventions): add the chat-links reference by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/10

## [1.1.0](https://github.com/jhoblitt/conventions-claude/compare/v1.0.0...v1.1.0) (2026-09-04)

### Features

* **code-conventions:** add a canon for the language-agnostic rules ([bb067fd](https://github.com/jhoblitt/conventions-claude/commit/bb067fdc27324f23fb487ba6b93f0900168bcbfe))


## What's Changed
* feat: add the code-conventions plugin by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/4

## 1.0.0 (2026-09-04)

### Features

* add the marketplace skeleton ([7001dec](https://github.com/jhoblitt/conventions-claude/commit/7001dec13d0e867b70147c72d9ae88ffaba6ec22))
* **github-conventions:** add the canon skill and references ([320af7b](https://github.com/jhoblitt/conventions-claude/commit/320af7b4146b4907725c458a405060cbe3ebd4a0))
* **github-conventions:** add the converge and new-repo skills ([d569e00](https://github.com/jhoblitt/conventions-claude/commit/d569e008117412694d327e1967c9d0b844dd0e01))
* **github-conventions:** add the ghconv-audit tool ([0a6200d](https://github.com/jhoblitt/conventions-claude/commit/0a6200df501d4cb5bffd542793644119f9061e9c))
* **go-conventions:** add goconv-audit, go-migrator, and go-converge ([0e4f105](https://github.com/jhoblitt/conventions-claude/commit/0e4f105b598d8e3b9900dc35c331852a4d43aec0))
* **go-conventions:** add the canon skill and references ([08db7a3](https://github.com/jhoblitt/conventions-claude/commit/08db7a3bb362ece234d89fa552b2231d7f008590))
* **go-conventions:** add the Go review canon ([90a4084](https://github.com/jhoblitt/conventions-claude/commit/90a4084deb6e9549b2042178eee9196032736e2d))
* **go-conventions:** add the go-new-project scaffold skill ([2df7ca8](https://github.com/jhoblitt/conventions-claude/commit/2df7ca801e87b85c743173f09f550423b2a6a6a6))
* **go-conventions:** add the go-review skill and go-reviewer agent ([6546712](https://github.com/jhoblitt/conventions-claude/commit/65467122f704fd0004892360fca1bb891651a7ac))
* **go-conventions:** add the goconv-hook formatter and session pointer ([8d3f36b](https://github.com/jhoblitt/conventions-claude/commit/8d3f36bb7fc5b75c0364ac460d164e3bc1575ca4))
* **go-conventions:** add the scaffold templates ([167a760](https://github.com/jhoblitt/conventions-claude/commit/167a760127d2c081f194592c608d5487cd28785e))

### Performance Improvements

* **go-conventions:** shorten the agent descriptions ([3428441](https://github.com/jhoblitt/conventions-claude/commit/342844165ddbfa006949bf790decbe48038a69e7))
* **skills:** move the audit contracts behind pointers ([4601a03](https://github.com/jhoblitt/conventions-claude/commit/4601a03c931338f2cd225915ea272575ccf2c138))

### Refactoring

* **github-conventions:** make the breaking-footer check a Go file ([0c22a89](https://github.com/jhoblitt/conventions-claude/commit/0c22a896fcdf64f796f6d1beea2d3af497fc12da))

### Documentation

* describe both plugins in the README ([5739352](https://github.com/jhoblitt/conventions-claude/commit/57393523ba3c8fe15439eadf2630b36e4da184a2))


## What's Changed
* feat: add the go-conventions and github-conventions plugins by @jhoblitt in https://github.com/jhoblitt/conventions-claude/pull/1

## New Contributors
* @jhoblitt made their first contribution in https://github.com/jhoblitt/conventions-claude/pull/1
