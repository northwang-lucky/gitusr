# Changelog

## [1.0.1](https://github.com/northwang-lucky/gitusr/compare/v1.0.0...v1.0.1) (2026-06-01)


### Bug Fixes

* **release:** 修复 brew install 后无法使用 gu 别名 ([6e8e233](https://github.com/northwang-lucky/gitusr/commit/6e8e2334d692de73187c726221b84771283500b8))
* **release:** 修复 brew install 后无法使用 gu 别名的问题 ([39facc8](https://github.com/northwang-lucky/gitusr/commit/39facc8010feb7eaf948666e1a0887ac462e1206))

## [1.0.0](https://github.com/northwang-lucky/gitusr/compare/v0.2.1...v1.0.0) (2026-05-31)


### ⚠ BREAKING CHANGES

* 项目达到稳定状态，API 和 CLI 接口已确定，发布 1.0.0 正式版本。

### Features

* 正式发布 v1.0.0 ([c6d6a60](https://github.com/northwang-lucky/gitusr/commit/c6d6a60bddf6ef1f633d422803a823cd19757e2e))

## [0.2.1](https://github.com/northwang-lucky/gitusr/compare/v0.2.0...v0.2.1) (2026-05-31)


### Bug Fixes

* **hook:** git clone wrapper 总是调用 gitusr use 并回到原目录 ([ab2eb89](https://github.com/northwang-lucky/gitusr/commit/ab2eb89359133558c0b58c69590ff02fdc2e1fe6))
* **hook:** git clone wrapper 总是调用 gitusr use 并回到原目录 ([85f036d](https://github.com/northwang-lucky/gitusr/commit/85f036daccf4d4d6fc4f5ff604d5b09a3cb5114f))

## [0.2.0](https://github.com/northwang-lucky/gitusr/compare/v0.1.1...v0.2.0) (2026-05-31)


### Features

* **cli:** 为 hook install 和 uninstall 实现 --all 标志 ([9e94337](https://github.com/northwang-lucky/gitusr/commit/9e9433777084d2ef2c266d6d7b103b0144eacddc))
* **cli:** 新增隐藏的 hook apply-rc 子命令 ([8d331ff](https://github.com/northwang-lucky/gitusr/commit/8d331ffcf8e692c1a895ae8ae7c451d9d130c31c))
* **hook:** CLI env 命令 ([a00c899](https://github.com/northwang-lucky/gitusr/commit/a00c899eb148a960624520bf34b8d9628f08d533))
* **hook:** CLI install/uninstall 命令 ([83447e3](https://github.com/northwang-lucky/gitusr/commit/83447e3da8c866168d2b0c18fe629971a867b7cb))
* **hook:** hook env 协调器和 apply-rc 逻辑 ([08a8bc1](https://github.com/northwang-lucky/gitusr/commit/08a8bc1bdaaf73b4a785e003875f8ae6b0f9f34a))
* **hook:** hook uninstaller 逻辑 ([c9c740c](https://github.com/northwang-lucky/gitusr/commit/c9c740c8592c70210c800d0bced64024a36d1989))
* **hook:** hook 状态持久化 ([65c9a65](https://github.com/northwang-lucky/gitusr/commit/65c9a654e9df7879ccc9ecc1b8e1a176ecdeafeb))
* **hook:** i18n 中英文翻译消息 ([16be415](https://github.com/northwang-lucky/gitusr/commit/16be4159f31960b50a5d3b36b1312851f6a4999e))
* **hook:** shell config 文件写入和标记块管理 ([f6784c5](https://github.com/northwang-lucky/gitusr/commit/f6784c5cd2c7a9e5dda72effe37e4d40c9d981d2))
* **hook:** 在 hook 命令下注册 env 子命令 ([ee4d85e](https://github.com/northwang-lucky/gitusr/commit/ee4d85ee67ce732a5b434b240e5436c0876fdd09))
* **hook:** 定义 domain 类型和 RC 文件解析与匹配 ([29ca334](https://github.com/northwang-lucky/gitusr/commit/29ca334f582cd2e3844477b61abb5ce92f7c2cbc))
* **hook:** 实现 bash cd auto-switch env 代码生成 ([c10c1b4](https://github.com/northwang-lucky/gitusr/commit/c10c1b47223501f061c64a672fb8c4c102ef73c8))
* **hook:** 实现 bash shell wrapper 代码生成 ([3928b3f](https://github.com/northwang-lucky/gitusr/commit/3928b3f90dab6ceb6294506fd0afeb4d04aea833))
* **hook:** 实现 zsh chpwd auto-switch env 代码生成 ([bf3f526](https://github.com/northwang-lucky/gitusr/commit/bf3f526a72accbb8da7a65906c445c30adbf7af0))
* **hook:** 实现 zsh shell wrapper 代码生成 ([487f5df](https://github.com/northwang-lucky/gitusr/commit/487f5dfcea8f4baf0b01767cf7094b262fb65104))
* **hook:** 将 hook env 重构为 install --type=cd 并支持 uninstall ([68a1de1](https://github.com/northwang-lucky/gitusr/commit/68a1de109dcea66b480da2c7ed57edda63457149))
* **hook:** 添加 AllHookTypes 有序切片到 types.go ([d2ca87d](https://github.com/northwang-lucky/gitusr/commit/d2ca87dc2a5a5755b31b41d5d3d7b5ea264e0a69))
* **use:** 添加 --silent-if-unchanged 标志支持静默切换 ([1d260f3](https://github.com/northwang-lucky/gitusr/commit/1d260f3ca80b14aeeb0afbe556ef0a9fb22808ac))


### Bug Fixes

* **goreleaser:** change homebrew tap directory from Formula to Casks ([40b7c47](https://github.com/northwang-lucky/gitusr/commit/40b7c4751adc03ebd6a0a6185e36129b378845fd))
* **hook:** 修复 Final Verification Wave 发现的三个阻塞问题 ([487cb13](https://github.com/northwang-lucky/gitusr/commit/487cb130b9eef5df6af132d55c44c5bac5a732a4))
* **hook:** 修复 uninstaller_test.go 中未使用的 fmt import 和 marker 引用 ([326cb73](https://github.com/northwang-lucky/gitusr/commit/326cb73658a2d241cd680e1412c6bae1cb55c6d1))
* **hook:** 修复 wrapper 在无参数时进入交互模式的问题 ([46224e0](https://github.com/northwang-lucky/gitusr/commit/46224e012da7489fa1368f9bf899bd97cd87611d))
* **hook:** 将 env 脚本中不存在的 count-users 替换为 list | wc -l ([10e8c67](https://github.com/northwang-lucky/gitusr/commit/10e8c67745e575ab044fb243649695c10b74a375))

## [0.1.1](https://github.com/northwang-lucky/gitusr/compare/v0.1.0...v0.1.1) (2026-05-31)


### Bug Fixes

* **cli:** 修复 list 和 remove 命令的错误重复打印问题 ([b05d719](https://github.com/northwang-lucky/gitusr/commit/b05d719d8cce51462ccf0dcc9d1cda0bc7ef05f6))
* **cli:** 修复 list 和 remove 命令的错误重复打印问题 ([b05d719](https://github.com/northwang-lucky/gitusr/commit/b05d719d8cce51462ccf0dcc9d1cda0bc7ef05f6))
* **cli:** 修复 list 和 remove 命令的错误重复打印问题 ([e5f20da](https://github.com/northwang-lucky/gitusr/commit/e5f20da9b30a40d248df7140d07604916ba75e67))
## v1.0.0
