# Changelog

## [0.6.13](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.12...v0.6.13) (2026-05-21)


### Bug Fixes

* package release image with scratch copy build ([#63](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/63)) ([60c6784](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/60c67849625103cd89e030063507e16b0d78c145))

## [0.6.12](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.11...v0.6.12) (2026-05-21)


### Bug Fixes

* tolerate static files in release packaging ([#61](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/61)) ([25526c0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/25526c04fd314e2c5465e88e0eb453a3b3d8bfdc))

## [0.6.11](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.10...v0.6.11) (2026-05-21)


### Bug Fixes

* package release images without docker start ([#59](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/59)) ([0a42e13](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/0a42e13ec38b39eb4496c54a9cf5190f7bf40ccd))

## [0.6.10](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.9...v0.6.10) (2026-05-21)


### Bug Fixes

* build release images via docker commit ([59fc5d2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/59fc5d21de9197c05b7e1e5f37160ac396f935c1))
* build release images via docker commit ([4f693ca](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/4f693ca0e47edffc5bfef0400815fe9d9dc78759))

## [0.6.9](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.8...v0.6.9) (2026-05-21)


### Bug Fixes

* use buildkit docker driver for releases ([d965bfc](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/d965bfc08db3ccde6f7261dd10db7273c9565ac8))
* use buildkit docker driver for releases ([e16719b](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e16719bfac8db096192b9c8d0b572bf68ad93ea2))

## [0.6.8](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.7...v0.6.8) (2026-05-21)


### Bug Fixes

* repair dev null before apt in lxc builds ([e2dca12](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e2dca125ef05d3c1fba0f3278f412cabcb952edc))
* repair dev null before apt in lxc builds ([77e3187](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/77e3187434b7affa440dc9eb880883c6f64426d2))

## [0.6.7](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.6...v0.6.7) (2026-05-21)


### Bug Fixes

* use supported Ubuntu base for Vulkan release ([f2a1c06](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/f2a1c0655576ae3e68fd95eebac0352d736c160c))
* use supported Ubuntu base for Vulkan release ([7d2bf44](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/7d2bf449102b8b4a693ae41c571f2519f783f379))

## [0.6.6](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.5...v0.6.6) (2026-05-21)


### Bug Fixes

* prebuild wrapper before docker lxc image builds ([c64038d](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/c64038dcd1bf83c1c0c517d3823a8bd0004048d8))

## [0.6.5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.4...v0.6.5) (2026-05-21)


### Bug Fixes

* set go env in docker lxc builds ([62f27d7](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/62f27d7439475e15cd8160d4f733afdd7ee08f69))

## [0.6.4](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.3...v0.6.4) (2026-05-21)


### Bug Fixes

* call go directly in docker lxc builds ([eb338d5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/eb338d56e8b8c44d2ed0e2472eda4b5004734b5e))

## [0.6.3](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.2...v0.6.3) (2026-05-21)


### Bug Fixes

* publish releases with docker cli ([613e6da](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/613e6dab79eb26ecce67b0d5cde239dfa6d76e73))

## [0.6.2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.1...v0.6.2) (2026-05-21)


### Bug Fixes

* avoid race detector on self-hosted release ([#41](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/41)) ([1d6bb15](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/1d6bb153873275a11b2cb3fba448e38a07018a5f))

## [0.6.1](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.0...v0.6.1) (2026-05-21)


### Bug Fixes

* set self-hosted runner Go cache ([#39](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/39)) ([c4ac4d5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/c4ac4d5a4aa274a9f853c933c69b72c568724cad))

## [0.6.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.5.0...v0.6.0) (2026-05-21)


### Features

* download MTP head GGUF ([#36](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/36)) ([39ca3e5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/39ca3e554f10e93dd5c27dff7aa67bd79705a9f1))


### Bug Fixes

* use self-hosted release runners ([#38](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/38)) ([fb5392f](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/fb5392fbc8c82e4ca35855f97c98bd4afec23e2d))

## [0.5.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.4.3...v0.5.0) (2026-05-21)


### Features

* add Gemma4 MTP TurboQuant support ([#34](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/34)) ([5c34171](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/5c3417116ddd025d560bdd3398fe19e04ee2e590))

## [0.4.3](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.4.2...v0.4.3) (2026-05-18)


### Bug Fixes

* avoid redundant model downloads ([#31](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/31)) ([357ac20](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/357ac203d858db4f63c1fbe687ad9ee71ea98e00))

## [0.4.2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.4.1...v0.4.2) (2026-05-18)


### Bug Fixes

* generalize llama plugin description ([b665833](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/b6658336d20223c40ef70b7dd487063adb3779be))

## [0.4.1](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.4.0...v0.4.1) (2026-05-18)


### Bug Fixes

* pass llama memory safety args ([e830f19](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e830f1932617e4a5e09bbd3db82f00fd65d0a9d4))
* pass llama memory safety args ([d497792](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/d49779201da431a62bf1e6167dff188bfe3cff45))

## [0.4.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.3.0...v0.4.0) (2026-05-18)


### Features

* add install-time GPU selection ([66ae19f](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/66ae19f1358a805f3dafedc8953803d0f100f7c3))
* add install-time GPU selection ([82927b9](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/82927b91634a3f8d903b258b69dad16bc15a58f9))

## [0.3.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.5...v0.3.0) (2026-05-17)


### Features

* show model download progress ([#23](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/23)) ([4001500](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/4001500cc25071289069e782a614dba3a5306276))

## [0.2.5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.4...v0.2.5) (2026-05-17)


### Bug Fixes

* surface model URL errors in plugin UI ([#21](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/21)) ([665c020](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/665c020425957d88453d6dc0898b174e7c1dd934))

## [0.2.4](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.3...v0.2.4) (2026-05-17)


### Bug Fixes

* handle Hugging Face browser model URLs ([#19](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/19)) ([717be1e](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/717be1eda00a2b610e5fb48193a234176da745fd))

## [0.2.3](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.2...v0.2.3) (2026-05-17)


### Bug Fixes

* dispatch release publisher with explicit repo ([6f18aec](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/6f18aec18e1af861c86ab65fb6e497182db77435))
* dispatch release publisher with explicit repo ([7c98337](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/7c98337e2fc49581adc6029d1c05dd52f7bc599a))

## [0.2.2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.1...v0.2.2) (2026-05-17)


### Bug Fixes

* dispatch release publisher after release-please ([abf53e2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/abf53e2281e6c213dd96ffbba0303886f75ba3df))
* dispatch release publisher after release-please ([d51fe4f](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/d51fe4fd391a8edf52adb41520454f052ae8b80f))

## [0.2.1](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.2.0...v0.2.1) (2026-05-17)


### Bug Fixes

* render plugin image digests separately ([c884b20](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/c884b20b6e8d77cf9d42e41e816b0a35281748e9))
* render plugin image digests separately ([f378e20](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/f378e20d92c32187971c19ccb16ff2283d26a5ed))

## [0.2.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.1.0...v0.2.0) (2026-05-17)


### Features

* add llama temperature config ([e14f808](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e14f8084aa61a42f7f300a8f397ae37e0dd35fb7))
* download configured model url in container ([#12](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/12)) ([8e4fb4d](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/8e4fb4db0ccfe637976c79b9200ced3eeab38dff))
