# gitusr

> 一个用于管理和切换 Git 用户身份的 CLI 工具。
>
> [English Documentation](./README_EN.md)

---

## 解决了什么问题？

在日常开发中，开发者经常需要在**个人项目**和**公司项目**之间切换不同的 Git 身份：

- 个人开源项目使用 `个人邮箱`
- 公司内部项目使用 `公司邮箱`

手动修改 `git config user.name` 和 `git config user.email` 既繁琐又容易出错。`gitusr` 帮你把常用的 Git 用户身份保存起来，通过一条命令就能在仓库级别或全局快速切换，无需记忆复杂的 `git config` 语法。

此外，`gitusr` 还提供**历史记录作者替换**功能，当你发现历史提交中的作者信息有误时，可以安全地重写历史。

## 安装指南

### 使用 Homebrew / LinuxBrew 安装（推荐）

```bash
brew tap northwang-lucky/gitusr
brew install gitusr
```

安装完成后，Homebrew 会自动将 `gitusr` 和快捷命令 `gu` 放入你的 `PATH` 中。

## 快速上手

以下是一个**最短路径 Demo**，演示如何用 `gitusr` 管理你的 Git 用户：

```bash
# 1. 初始化 —— 从当前 git 全局配置导入第一个用户
#    （如果全局没有配置，会交互式提示输入）
gitusr init

# 2. 添加另一个用户（例如公司身份）
gitusr add
# 提示输入 user.name:  Zhang San
# 提示输入 user.email: zhangsan@company.com

# 3. 查看已保存的所有用户
gitusr list
# 输出：
# 0：姓名：North Wang       | 邮箱：north@personal.com
# 1：姓名：Zhang San        | 邮箱：zhangsan@company.com

# 4. 切换到公司用户（在当前仓库中）
cd ~/work/company-project
gitusr use --index 1
# 成功！
# 您的 repo git 用户为：
#
# user.name  = Zhang San
# user.email = zhangsan@company.com

# 5. 查看当前仓库使用的用户
gitusr current

# 6. 也可以直接用索引、姓名或邮箱切换
gitusr use --email zhangsan@company.com
gitusr use --name "Zhang San"
```

## CLI 文档

### 全局说明

- 所有命令均可通过 `gitusr <command> --help` 查看帮助。
- 同时支持 `gu` 作为快捷命令（例如 `gu list`）。
- 数据文件默认保存在 `$XDG_DATA_HOME/gitusr/user-list.json`（回退到 `~/.local/share/gitusr/user-list.json`）。
- 支持中英文双语，语言优先级：`GITUSR_LANG` > `LANGUAGE` > `LANG` > `en`。

### `gitusr init` — 从 git 全局配置初始化

从当前 Git 全局配置读取 `user.name` 和 `user.email`，保存为第一个用户。如果检测到旧版配置文件（`~/.gitusr/user-list.json`），会提示迁移到 XDG 目录。

**用法：**

```bash
gitusr init [flags]
```

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--force` | `-f` | 强制覆盖已存在的用户列表 |
| `--name` | `-n` | 非交互式指定用户名（必须与 `--email` 同时使用） |
| `--email` | `-e` | 非交互式指定用户邮箱（必须与 `--name` 同时使用） |
| `--yes` | `-y` | 跳过所有确认提示 |

**示例：**

```bash
# 交互式初始化
gitusr init

# 强制重新初始化
gitusr init --force

# 非交互式直接指定
gitusr init --name "North Wang" --email "north@example.com"
```

---

### `gitusr add` — 添加并保存用户

交互式添加一个新的 Git 用户身份。支持通过 flag 非交互式添加。

**用法：**

```bash
gitusr add [flags]
```

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--name` | `-n` | 指定用户名（非交互式，必须与 `--email` 同时使用） |
| `--email` | `-e` | 指定用户邮箱（非交互式，必须与 `--name` 同时使用） |

**示例：**

```bash
# 交互式添加
gitusr add
# 提示输入 user.name:  North Wang
# 提示输入 user.email: north@example.com

# 非交互式添加
gitusr add --name "North Wang" --email "north@example.com"
```

---

### `gitusr use` — 切换 Git 用户

在当前 Git 仓库或全局范围内切换已保存的用户。支持通过索引、姓名或邮箱定位用户；如果未指定且只有一个用户，直接切换；如果有多个用户，会进入交互式选择。

**用法：**

```bash
gitusr use [flags]
```

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--global` | `-g` | 切换全局 Git 用户（而非当前仓库） |
| `--name` | `-n` | 按姓名切换 |
| `--email` | `-e` | 按邮箱切换 |
| `--index` | `-i` | 按索引切换（通过 `gitusr list` 查看索引） |

**示例：**

```bash
# 在当前仓库切换（交互式选择）
gitusr use

# 按索引切换
gitusr use --index 1

# 按邮箱切换
gitusr use --email north@example.com

# 全局切换
gitusr use --global --name "North Wang"
```

---

### `gitusr current` — 显示当前用户

显示当前仓库级别或全局的 Git 用户配置。

**别名：** `ct`

**用法：**

```bash
gitusr current [flags]
```

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--global` | `-g` | 显示全局 Git 用户 |

**示例：**

```bash
# 查看当前仓库用户
gitusr current

# 查看全局用户
gitusr current --global
```

---

### `gitusr list` — 显示所有已保存用户

列出所有已保存的 Git 用户身份，每行显示索引、姓名和邮箱。

**别名：** `ls`

**用法：**

```bash
gitusr list
```

**示例：**

```bash
gitusr list
# 0：姓名：North Wang       | 邮箱：north@personal.com
# 1：姓名：Zhang San        | 邮箱：zhangsan@company.com
```

---

### `gitusr remove` — 删除用户

删除一个已保存的 Git 用户身份。支持通过索引、姓名或邮箱定位。

**别名：** `rm`

**用法：**

```bash
gitusr remove [flags]
```

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--name` | `-n` | 按姓名删除 |
| `--email` | `-e` | 按邮箱删除 |
| `--index` | `-i` | 按索引删除 |

**示例：**

```bash
# 按索引删除
gitusr remove --index 1

# 按邮箱删除
gitusr remove --email zhangsan@company.com
```

---

### `gitusr replace` — 替换 git 历史记录中的作者

使用 `git-filter-repo` 重写历史提交记录，将匹配指定邮箱的提交作者替换为另一个已保存用户。操作前会自动创建备份分支。

**前提条件：** 已安装 `git-filter-repo`（`pip install git-filter-repo`）

**用法：**

```bash
gitusr replace <target-email> [flags]
```

**参数：**

| 参数 | 说明 |
|------|------|
| `target-email` | 需要被替换的旧作者邮箱 |

**标志：**

| 标志 | 简写 | 说明 |
|------|------|------|
| `--with-name` | | 指定新用户的姓名 |
| `--with-email` | | 指定新用户的邮箱 |
| `--with-index` | | 指定新用户的索引 |
| `--yes` | `-y` | 历史重写后自动切换仓库用户，跳过确认 |

**示例：**

```bash
# 将所有提交中 old@wrong.com 的作者替换为索引 1 的用户
gitusr replace old@wrong.com --with-index 1

# 替换完成后自动切换仓库用户（跳过确认）
gitusr replace old@wrong.com --with-name "North Wang" --with-email "north@example.com" --yes
```

---

## 卸载

```bash
brew uninstall gitusr
```

卸载后，用户数据文件不会自动删除，如需清理可手动执行：

```bash
rm -rf "${XDG_DATA_HOME:-$HOME/.local/share}/gitusr"
```

## 许可证

[MIT](./LICENSE)
