# Changelog

## [0.7.36](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.35...v0.7.36) (2026-05-22)


### Bug Fixes

* run wrapper release tests without cgo

## [0.7.35](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.34...v0.7.35) (2026-05-22)


### Bug Fixes

* stabilize offline native release builds ([#143](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/143)) ([a4f32b9](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/a4f32b95c0c037779cb1134e12fd479f82daa460))
* verify baked llama source ref ([#145](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/145)) ([e835b53](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e835b5386a62628538abd3042a8e85e3ce755bf7))

## [0.7.34](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.33...v0.7.34) (2026-05-22)


### Bug Fixes

* require baked tools for gpu releases ([#142](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/142)) ([4ddb195](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/4ddb195cfc04e592a3cbe8c014c07558eafd2412))

## [0.7.33](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.32...v0.7.33) (2026-05-21)


### Bug Fixes

* build accelerator images natively on self-hosted runners ([#140](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/140)) ([1b1ac8c](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/1b1ac8c))

## [0.7.32](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.31...v0.7.32) (2026-05-21)


### Bug Fixes

* serialize accelerator image builds ([#138](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/138)) ([a54536c](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/a54536c))

## [0.7.31](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.30...v0.7.31) (2026-05-21)


### Bug Fixes

* stabilize self-hosted accelerator builds ([#135](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/135)) ([2afc073](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/2afc073c24757824eb349f39c810f739b846326c))

## [0.7.30](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.29...v0.7.30) (2026-05-21)


### Bug Fixes

* avoid cuda image rm cleanup ([#133](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/133)) ([0de77e3](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/0de77e315c401fadd617b0f7dc77856c997f77a1))

## [0.7.29](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.28...v0.7.29) (2026-05-21)


### Bug Fixes

* set self-hosted publish path and jobs ([#131](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/131)) ([de03fcc](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/de03fcc2f16894991f8f64f7fbb5b81560c7c13a))

## [0.7.28](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.27...v0.7.28) (2026-05-21)


### Bug Fixes

* respect runner build limits ([#129](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/129)) ([b3da718](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/b3da718023a558b6464aa544714eb73af618a5ba))

## [0.7.27](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.26...v0.7.27) (2026-05-21)


### Bug Fixes

* avoid incompatible CUDA cuBLAS dev package ([#127](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/127)) ([8073f39](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/8073f3970eec1c4fb9a1feca8d7ebb199674189d))

## [0.7.26](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.25...v0.7.26) (2026-05-21)


### Bug Fixes

* install CUDA dev packages for release image ([#125](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/125)) ([cbb10ef](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/cbb10ef830ff305c54199127e1180bee9d35adea))

## [0.7.25](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.24...v0.7.25) (2026-05-21)


### Bug Fixes

* include Ubuntu updates for CUDA compiler deps ([#123](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/123)) ([aaa1f58](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/aaa1f58bc753487edd4f85301400f3bf31d77cc3))

## [0.7.24](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.23...v0.7.24) (2026-05-21)


### Bug Fixes

* install host compiler for CUDA release image ([#121](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/121)) ([4cf29ac](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/4cf29aceb04af8794dd4fd66c855ec4d2d330ac3))

## [0.7.23](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.22...v0.7.23) (2026-05-21)


### Bug Fixes

* disable apt preconfigure in release images ([#119](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/119)) ([85bb046](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/85bb04699a8519c9d1ce73477d986bba30c77e37))

## [0.7.22](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.21...v0.7.22) (2026-05-21)


### Bug Fixes

* disable wrapper VCS stamping in release builds ([#117](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/117)) ([9f97995](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/9f979957d2cbbbf094cea9fa835796d5c970a76b))

## [0.7.21](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.20...v0.7.21) (2026-05-21)


### Bug Fixes

* bootstrap dpkg state for CUDA base images ([#115](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/115)) ([0724743](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/0724743c7f040e7bba37abfcfa68e5aab38dd902))

## [0.7.20](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.19...v0.7.20) (2026-05-21)


### Bug Fixes

* keep release control jobs on github runners ([#113](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/113)) ([29ffd40](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/29ffd40a34a65c667e63ce99243e09fea308d1dc))
* serialize CUDA source extraction ([#112](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/112)) ([56be20c](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/56be20c7347e6c1b41a170684147b0c21308b482))

## [0.7.19](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.18...v0.7.19) (2026-05-21)


### Bug Fixes

* extract CUDA source in one layer ([#110](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/110)) ([d45e789](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/d45e7895b202252ec84b12ec262621b1ba070c0d))

## [0.7.18](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.17...v0.7.18) (2026-05-21)


### Bug Fixes

* keep apt temp files off /tmp ([#108](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/108)) ([73d0efe](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/73d0efe27c0ec38f52bb1edd734c6b702d0077e6))

## [0.7.17](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.16...v0.7.17) (2026-05-21)


### Bug Fixes

* install CUDA NVVM tools from direct debs ([#106](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/106)) ([9a89b28](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/9a89b281b9d5f49323e5d539d3915305accd2759))

## [0.7.16](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.15...v0.7.16) (2026-05-21)


### Bug Fixes

* use CUDA base toolchain packages ([#104](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/104)) ([30f6c6b](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/30f6c6bf0ca8051d54a637d0641f6040db48362d))

## [0.7.15](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.14...v0.7.15) (2026-05-21)


### Bug Fixes

* install crane from pinned release archive ([#102](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/102)) ([bc68225](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/bc68225cea1c7c1fe12250af74fe876b3e6bd395))

## [0.7.14](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.13...v0.7.14) (2026-05-21)


### Bug Fixes

* stabilize CUDA apt and split publish runners ([#100](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/100)) ([f14a82c](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/f14a82cbf7b1ac9067a2ea9229cf3711fc60067b))

## [0.7.13](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.12...v0.7.13) (2026-05-21)


### Bug Fixes

* harden CUDA archive extraction ([#98](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/98)) ([3117c51](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/3117c510a875f031f92162081260306f2e3f05fa))

## [0.7.12](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.11...v0.7.12) (2026-05-21)


### Bug Fixes

* reinstall CUDA NVVM tools ([#96](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/96)) ([f54cc2e](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/f54cc2ec1421a4401425112e6a9d320738c292f0))

## [0.7.11](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.10...v0.7.11) (2026-05-21)


### Bug Fixes

* stabilize Dockerfile apt state ([#94](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/94)) ([060f2af](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/060f2afed0ce848ddd5c3aa0fde4cec1884f34d6))

## [0.7.10](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.9...v0.7.10) (2026-05-21)


### Bug Fixes

* install CUDA NVVM compiler bits ([#92](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/92)) ([295e245](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/295e245b4a8d9b1eaa5c39a0eb135325144e1157))

## [0.7.9](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.8...v0.7.9) (2026-05-21)


### Bug Fixes

* keep CUDA apt temp files off /tmp ([#90](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/90)) ([ee352ef](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/ee352ef82b1f84fffb09daf8222c43ce228e0f5c))

## [0.7.8](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.7...v0.7.8) (2026-05-21)


### Bug Fixes

* point CUDA Docker build at CMake modules ([#88](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/88)) ([9652f23](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/9652f237193b0e2e23e53de09d9465fb850736d1))

## [0.7.7](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.6...v0.7.7) (2026-05-21)


### Bug Fixes

* avoid shell arrays in Dockerfile cmake ([#86](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/86)) ([f0c5e09](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/f0c5e0909a5a5d57d4b4d5392d2be3b0322d26bd))

## [0.7.6](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.5...v0.7.6) (2026-05-21)


### Bug Fixes

* pass Docker archive URL as build arg ([#84](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/84)) ([24959ab](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/24959abefc28c1a47910477ff093a89cb72d15a1))

## [0.7.5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.4...v0.7.5) (2026-05-21)


### Bug Fixes

* fetch llama source as archive in Dockerfile ([#82](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/82)) ([d125ce4](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/d125ce40ae629992e4ed36cdeb034b9fe0c457c6))

## [0.7.4](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.3...v0.7.4) (2026-05-21)


### Bug Fixes

* create random devices in docker builds ([#80](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/80)) ([89f911b](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/89f911bf221a736ae3c70b38ebae9619667b3380))

## [0.7.3](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.2...v0.7.3) (2026-05-21)


### Bug Fixes

* build CUDA releases in CUDA image ([#78](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/78)) ([bd3b1fd](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/bd3b1fd323f4256d3170145982006b2acc88a239))

## [0.7.2](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.1...v0.7.2) (2026-05-21)


### Bug Fixes

* use buffered model copy fallback ([#76](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/76)) ([24e39ff](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/24e39ff8afbeebdeca93df4bcb670aea2d03723e))

## [0.7.1](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.7.0...v0.7.1) (2026-05-21)


### Bug Fixes

* install models across filesystems ([#74](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/74)) ([e49af5f](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e49af5f40a678f9e4669250ae110ef4980bb5e41))

## [0.7.0](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.16...v0.7.0) (2026-05-21)


### Features

* support Qwen3.6 NextN TurboQuant ([2f4a8ec](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/2f4a8ec827e5a34ddd64c01148222109a16c9b41))
* support Qwen3.6 NextN TurboQuant ([df763ca](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/df763cabbafbbd04ac54fcfa13ba0c04f0a7f609))

## [0.6.16](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.15...v0.6.16) (2026-05-21)


### Bug Fixes

* package Vulkan ICD libraries ([3c66368](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/3c6636884d6ec0d01cf0e8c58d20e814d0bd3959))
* package Vulkan ICD libraries ([37a4072](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/37a4072a1c366c3b93aa0f6b3ee235c40e13c1be))

## [0.6.15](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.14...v0.6.15) (2026-05-21)


### Bug Fixes

* build release backend libraries ([#68](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/68)) ([de6cbf5](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/de6cbf5c67907219ae6906d868e67e8d71605b17))
* package release rootfs contents ([#67](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/67)) ([df132bf](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/df132bf9093b1f273720aa760b19dae901b1b52d))

## [0.6.14](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/compare/v0.6.13...v0.6.14) (2026-05-21)


### Bug Fixes

* push release images with crane ([#65](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/issues/65)) ([e3bef1a](https://github.com/RakuenSoftware/smoothnas-plugin-llama-cpp/commit/e3bef1a50688d54cdddb4759721a7fc1f5e64e93))

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
