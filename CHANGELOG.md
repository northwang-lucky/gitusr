# Changelog

## [1.3.0](https://github.com/northwang-lucky/gitusr/compare/v1.2.0...v1.3.0) (2026-08-24)


### Features

* **hooks:** clone hook 支持按仓库 host 自动匹配用户 ([105829e](https://github.com/northwang-lucky/gitusr/commit/105829e3c0b3a783dd05fcee03233a7c73eb93c6))
* **hosts:** clone hook 自动按仓库 host 匹配 Git 用户 ([9974650](https://github.com/northwang-lucky/gitusr/commit/9974650652ef097abd1d104adcdd67b8d4ed29ca))
* **hosts:** 新增 host 路由规则模型、匹配逻辑与 JSON 存储 ([00f28dc](https://github.com/northwang-lucky/gitusr/commit/00f28dc7b4710fa6766d8765e4eb26183e39ea4f))


### Bug Fixes

* **test:** replace 测试在 git-filter-repo 不可用时自动降级 ([d3b273c](https://github.com/northwang-lucky/gitusr/commit/d3b273c4f579ee71057252d5b92cdf3478274125))

## [1.2.0](https://github.com/northwang-lucky/gitusr/compare/v1.1.0...v1.2.0) (2026-06-01)


### Features

* **cli:** hooks install 后提示用户 source shell rc 文件 ([85b2a35](https://github.com/northwang-lucky/gitusr/commit/85b2a3531694b32b2cffd0c72f10c8e3ba25a2e8))
* **cli:** hooks install 添加 source 提示并修复 CI 竞态 ([64682e4](https://github.com/northwang-lucky/gitusr/commit/64682e4ac7705ea2167b964b43157688a16afe4f))


### Bug Fixes

* **ci:** 合并 GoReleaser 到 Release Please 工作流以消除竞态条件 ([e35f37b](https://github.com/northwang-lucky/gitusr/commit/e35f37be2cff4acd29e029888ef0177210e48743))

## [1.1.0](https://github.com/northwang-lucky/gitusr/compare/v1.0.2...v1.1.0) (2026-06-01)


### Features

* **cli:** 新增 hooks apply-rc 子命令，从 hook apply-rc 移植 ([481bbf9](https://github.com/northwang-lucky/gitusr/commit/481bbf9b0f0b2a650c384064d048cb679c993ff2))
* **cli:** 添加 hooks 父命令和 is-disabled/enable/disable 子命令 ([af1495d](https://github.com/northwang-lucky/gitusr/commit/af1495d561faaa4f3461ac5daa56a001e665b769))
* **hooks:** add hooks disable command with positional arg ([585437d](https://github.com/northwang-lucky/gitusr/commit/585437de8d536fc62a8bb67cc0c2332df1de9402))
* **hooks:** 扩展 HookState 类型定义和状态管理 ([93f7072](https://github.com/northwang-lucky/gitusr/commit/93f707275ddd1c98030ca60ae87448d8cc5f8a25))
* **hooks:** 新增 EnableHook/DisableHook/IsEnabled 状态管理函数 ([bbf15f9](https://github.com/northwang-lucky/gitusr/commit/bbf15f980b54f0d68203377fa012e53b4d275886))
* **hooks:** 新增 GenerateUnifiedBashWrapper 统一 bash wrapper ([7600a1d](https://github.com/northwang-lucky/gitusr/commit/7600a1d5adf0dd5e7f6a460732a091c79c0f407e))
* **hooks:** 新增 hooks enable 命令，使用位置参数替代 --type 标志 ([e5abad9](https://github.com/northwang-lucky/gitusr/commit/e5abad9e5dd8f53e11112f9e96f4780ae62b76f0))
* **hooks:** 新增 hooks uninstall 命令并更新 uninstallFunc ([e65354c](https://github.com/northwang-lucky/gitusr/commit/e65354c90e2f99ca10d984ef1f39f2fdce12ad69))
* **hooks:** 重构 hook 命令体系并新增 hooks 子命令 ([93e9bd6](https://github.com/northwang-lucky/gitusr/commit/93e9bd6a6418ead24ccff6179ea9cd0d80f2fd2d))
* **hook:** 新增 CD 标记辅助函数 AppendCDSourceLine、RemoveCDSourceBlock 和 removeMarkedCDBlock ([2e46b9e](https://github.com/northwang-lucky/gitusr/commit/2e46b9e44c808e99a531195d27502fbf333135f2))
* **hook:** 新增 GenerateUnifiedZshWrapper 统一 zsh 封装器 ([7a95203](https://github.com/northwang-lucky/gitusr/commit/7a9520388339b6b363294e19d1477d4d40a68ae5))
* **hook:** 新增 UninstallAll 函数用于一键卸载所有 hook 类型 ([ae05c38](https://github.com/northwang-lucky/gitusr/commit/ae05c388d3a7536f071d2dc521a38b7b274d11ff))
* **hook:** 添加 InstallAll 以通过统一包装器安装所有 hook 类型 ([0924d14](https://github.com/northwang-lucky/gitusr/commit/0924d144d99e19f1754efea9b4b5e5bd327d4d59))
* **scripts:** 支持 GITUSR_E2E_SHELL 环境变量选择 zsh E2E 测试 ([7f01388](https://github.com/northwang-lucky/gitusr/commit/7f01388209a94440f596b89df57e971c15cb063f))
* **skills:** 新增二进制发布流程 skill ([0a1c854](https://github.com/northwang-lucky/gitusr/commit/0a1c8542dd911ad44ec3dfc1f090108d88541b94))


### Bug Fixes

* **cli,hook:** 修复 Final Verification Wave 中的关键问题 (F1, F4) ([712fcca](https://github.com/northwang-lucky/gitusr/commit/712fccabb511f1b9f55618ea2555f6a9265c9074))
* **hook:** CD 钩子使用 AppendCDSourceLine，避免与 clone/commit 钩子冲突 ([57dea8e](https://github.com/northwang-lucky/gitusr/commit/57dea8e9d6d4a34c1b1a90fce2294376b6572403))
* **hook:** deleteWrapperFiles 现在也删除 cd-env.{sh,zsh} ([d072a33](https://github.com/northwang-lucky/gitusr/commit/d072a33b5efcb44f2a2c9e89014e7aa1d10032b0))

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
