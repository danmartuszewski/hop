# Changelog

## [1.1.0](https://github.com/danmartuszewski/hop/compare/v1.0.1...v1.1.0) (2026-02-17)


### Features

* add MCP server for AI assistant integration ([93f7684](https://github.com/danmartuszewski/hop/commit/93f768455f5ae1b616c929812529680a6f6aa390))

## [1.0.1](https://github.com/danmartuszewski/hop/compare/v1.0.0...v1.0.1) (2026-02-16)


### Bug Fixes

* use PAT in release-please to trigger release workflow ([1b7a31e](https://github.com/danmartuszewski/hop/commit/1b7a31ea7116227cba03c59e5166a3169a0687aa))

## 1.0.0 (2026-02-15)


### Features

* add connection details panel to dashboard ([ce39d66](https://github.com/danmartuszewski/hop/commit/ce39d669db034ed16401b9964b04c122dce31e41))
* add context-aware SSH error suggestions ([21c972a](https://github.com/danmartuszewski/hop/commit/21c972a393d367a05e9901da152e085afd54c6a9))
* add export connections to YAML ([#4](https://github.com/danmartuszewski/hop/issues/4)) ([baa6428](https://github.com/danmartuszewski/hop/commit/baa642802956b0f7f200c96c447fe0103acd16f9))
* add GoReleaser and Homebrew tap distribution ([f76fc39](https://github.com/danmartuszewski/hop/commit/f76fc3961e441c518e8e9b0c4b185a24fd6433cd))
* add hop CLI tool for SSH connection management ([f2c5f71](https://github.com/danmartuszewski/hop/commit/f2c5f7179b92fc9c447f256db81bf664c56471dc))
* add mouse support to dashboard ([6ec0254](https://github.com/danmartuszewski/hop/commit/6ec0254382eb9744f8ad70046e930aa3ba7c53ed))
* add recent connections tracking ([c94b495](https://github.com/danmartuszewski/hop/commit/c94b495e4393d2d2ecfc0b185667476708b54572))
* add shell completion support ([9ffd7e4](https://github.com/danmartuszewski/hop/commit/9ffd7e4fcb5dd47da5746e88a03766598e909849))
* add SSH config import with ProxyJump and ForwardAgent support ([305530c](https://github.com/danmartuszewski/hop/commit/305530ce5f8e972ca40996bedab8cbbcbdf63aad))
* add tag filter UI to dashboard ([9bf9ad0](https://github.com/danmartuszewski/hop/commit/9bf9ad057a1bb0475ddadcdbd308968a713d2aa2))
* implement multi-exec command for parallel SSH execution ([a920226](https://github.com/danmartuszewski/hop/commit/a9202262f4546d4aa5bbfeeaa48927247dff0575))
* implement terminal detection for multi-open feature ([52d61c6](https://github.com/danmartuszewski/hop/commit/52d61c655559dbbe2d224375e5961fde8402b135))
* return to dashboard after SSH session ends ([b4680d2](https://github.com/danmartuszewski/hop/commit/b4680d26b7a20271814e755cf3409145b9399b84)), closes [#7](https://github.com/danmartuszewski/hop/issues/7)
* support multi-keyword filter with AND logic ([2d9b26a](https://github.com/danmartuszewski/hop/commit/2d9b26adbf4626ae346fa403d7cfd2b460295371))
* use fuzzy matching with scoring in dashboard filter ([7091968](https://github.com/danmartuszewski/hop/commit/709196809d1ace9c331fc13d24917865fec112a4))


### Bug Fixes

* improve connection form UX ([5de409d](https://github.com/danmartuszewski/hop/commit/5de409d7c3be898b0c78dfd5dfaba95f35e10f62))
* improve port display in connection list ([8a75aad](https://github.com/danmartuszewski/hop/commit/8a75aad99d815b9c913b1225c82515fac16fe758))
* prevent config save from adding empty values for omitted keys ([4d05fda](https://github.com/danmartuszewski/hop/commit/4d05fdae25b0521c162e6b5ed1c184c1d4f22d2b))
* remove truncated placeholders from edit form fields ([ad063f6](https://github.com/danmartuszewski/hop/commit/ad063f6d39a5d7de3b35abd6bf1d62c31c976a4d))
* **tui:** align mouse click hit-testing with rendered rows ([a554022](https://github.com/danmartuszewski/hop/commit/a5540226f39afaaa77f45926bad999ad16ac5df0))
