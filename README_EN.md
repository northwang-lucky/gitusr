# gitusr

> A CLI tool for managing and switching between Git user identities.
>
> [中文文档](./README.md)

---

## What Problem Does It Solve?

Developers frequently need to switch between different Git identities for **personal projects** and **work projects**:

- Use `personal email` for open-source projects
- Use `company email` for internal company projects

Manually editing `git config user.name` and `git config user.email` is tedious and error-prone. `gitusr` lets you save your commonly used Git identities and switch between them at the repository or global level with a single command — no need to memorize complex `git config` syntax.

Additionally, `gitusr` provides:

- **History author rewriting**: safely fix incorrect author information in past commits
- **Shell Hook auto-switch**: automatically detect `.gitusrrc` configuration and apply the corresponding Git user identity during `git clone`, `git commit`, `cd`, and other operations

## Installation

### Install via Homebrew / LinuxBrew (recommended)

```bash
brew tap northwang-lucky/gitusr
brew install gitusr
```

After installation, Homebrew automatically places both `gitusr` and the `gu` shortcut in your `PATH`.

## Quick Start

Here is a **shortest-path demo** showing how to manage your Git users with `gitusr`:

```bash
# 1. Initialize — import the first user from your existing git global config
#    (If no global config exists, you will be prompted interactively)
gitusr init

# 2. Add another user (e.g., your work identity)
gitusr add
# Prompt for user.name:  Zhang San
# Prompt for user.email: zhangsan@company.com

# 3. List all saved users
gitusr list
# Output:
# 0: Name: North Wang       | Email: north@personal.com
# 1: Name: Zhang San        | Email: zhangsan@company.com

# 4. Switch to the work user (in the current repository)
cd ~/work/company-project
gitusr use --index 1
# Success!
# Your repo git user is:
#
# user.name  = Zhang San
# user.email = zhangsan@company.com

# 5. Check which user is active in the current repo
gitusr current

# 6. You can also switch by email or name directly
gitusr use --email zhangsan@company.com
gitusr use --name "Zhang San"
```

## CLI Reference

### Global Notes

- Use `gitusr <command> --help` to view help for any command.
- The `gu` alias is also supported (e.g., `gu list`).
- User data is stored at `$XDG_DATA_HOME/gitusr/user-list.json` (falls back to `~/.local/share/gitusr/user-list.json`).
- Supports English and Chinese. Locale priority: `GITUSR_LANG` > `LANGUAGE` > `LANG` > `en`.

### `gitusr init` — Initialize from git global config

Reads `user.name` and `user.email` from your global Git config and saves them as the first user. If a legacy config file (`~/.gitusr/user-list.json`) is detected, you will be prompted to migrate it to the XDG directory.

**Usage:**

```bash
gitusr init [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Force overwrite existing user list |
| `--name` | `-n` | Specify user name non-interactively (must be used with `--email`) |
| `--email` | `-e` | Specify user email non-interactively (must be used with `--name`) |
| `--yes` | `-y` | Skip all confirmation prompts |

**Examples:**

```bash
# Interactive initialization
gitusr init

# Force re-initialization
gitusr init --force

# Non-interactive
gitusr init --name "North Wang" --email "north@example.com"
```

---

### `gitusr add` — Add and save a user

Interactively add a new Git user identity. Also supports non-interactive addition via flags.

**Usage:**

```bash
gitusr add [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Specify user name (non-interactive, must be used with `--email`) |
| `--email` | `-e` | Specify user email (non-interactive, must be used with `--name`) |

**Examples:**

```bash
# Interactive add
gitusr add
# Prompt for user.name:  North Wang
# Prompt for user.email: north@example.com

# Non-interactive add
gitusr add --name "North Wang" --email "north@example.com"
```

---

### `gitusr use` — Switch Git user

Switch to a saved user in the current Git repository or globally. Supports locating the user by index, name, or email. If no filter is given and only one user exists, it switches directly; if multiple users exist, an interactive prompt is shown.

**Usage:**

```bash
gitusr use [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--global` | `-g` | Switch the global Git user instead of the repository-level user |
| `--name` | `-n` | Switch by name |
| `--email` | `-e` | Switch by email |
| `--index` | `-i` | Switch by index (see `gitusr list`) |

**Examples:**

```bash
# Switch in current repo (interactive selection)
gitusr use

# Switch by index
gitusr use --index 1

# Switch by email
gitusr use --email north@example.com

# Switch globally
gitusr use --global --name "North Wang"
```

---

### `gitusr current` — Show current user

Display the current repository-level or global Git user configuration.

**Alias:** `ct`

**Usage:**

```bash
gitusr current [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--global` | `-g` | Show the global Git user |

**Examples:**

```bash
# Show current repo user
gitusr current

# Show global user
gitusr current --global
```

---

### `gitusr list` — Show all saved users

List all saved Git user identities, showing index, name, and email per line.

**Alias:** `ls`

**Usage:**

```bash
gitusr list
```

**Example:**

```bash
gitusr list
# 0: Name: North Wang       | Email: north@personal.com
# 1: Name: Zhang San        | Email: zhangsan@company.com
```

---

### `gitusr remove` — Delete a user

Remove a saved Git user identity. Supports locating the user by index, name, or email.

**Alias:** `rm`

**Usage:**

```bash
gitusr remove [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Remove by name |
| `--email` | `-e` | Remove by email |
| `--index` | `-i` | Remove by index |

**Examples:**

```bash
# Remove by index
gitusr remove --index 1

# Remove by email
gitusr remove --email zhangsan@company.com
```

---

### `gitusr replace` — Replace author in git history

Rewrite commit history using `git-filter-repo`, replacing the author of commits matching `<target-email>` with another saved user. A backup branch is created automatically before any destructive operation.

**Prerequisite:** `git-filter-repo` installed (`pip install git-filter-repo`)

**Usage:**

```bash
gitusr replace <target-email> [flags]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `target-email` | The old author email to be replaced |

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--with-name` | | Specify the replacement user's name |
| `--with-email` | | Specify the replacement user's email |
| `--with-index` | | Specify the replacement user's index |
| `--yes` | `-y` | Automatically switch the repo user after rewrite, skipping confirmation |

**Examples:**

```bash
# Replace all commits authored by old@wrong.com with the user at index 1
gitusr replace old@wrong.com --with-index 1

# Replace and automatically switch repo user (skip confirmation)
gitusr replace old@wrong.com --with-name "North Wang" --with-email "north@example.com" --yes
```

---

### `gitusr hook` — Manage Shell Auto-Switch Hooks

After installing Shell hooks, `gitusr` can automatically detect and switch Git user identities during `git clone`, `git commit`, `cd`, and other operations.

**Usage:**

```bash
gitusr hook <subcommand> [flags]
```

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `install` | Install shell hooks |
| `uninstall` | Uninstall shell hooks |

#### `gitusr hook install`

Install auto-switch hooks for the current shell (bash and zsh). Supports three types:

- **`clone`** — Automatically enters the cloned repository and calls `gitusr use` to switch the user after `git clone`. Supports specifying the user via `--gu-name` / `--gu-email` arguments for non-TTY environments
- **`commit`** — Automatically reads `.gitusrrc` during `git commit` and applies the user
- **`cd`** — Automatically applies the user when `cd`ing into a directory containing `.gitusrrc`

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | | Hook type: `clone`, `commit`, or `cd` |
| `--all` | `-a` | Install all three hook types |

**Examples:**

```bash
# Install cd hook (auto-apply .gitusrrc when switching directories)
gitusr hook install --type cd

# Install clone hook (auto-switch user during git clone)
gitusr hook install --type clone

# Install all hooks
gitusr hook install --all
```

#### `gitusr hook uninstall`

Uninstall installed shell hooks.

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | | Hook type: `clone`, `commit`, or `cd` |
| `--all` | `-a` | Uninstall all hooks |

**Examples:**

```bash
# Uninstall cd hook
gitusr hook uninstall --type cd

# Uninstall all hooks
gitusr hook uninstall --all
```

#### `.gitusrrc` File

Create a `.gitusrrc` file in the root of a Git repository. When `cd`ing into that directory or running `git commit`, the hook will automatically match and apply the corresponding Git user.

**Format:**

```json
{
  "name": "Zhang San",
  "email": "zhangsan@company.com"
}
```

Matching priority: **email > name**. Only one field is required.

---

## Uninstall

```bash
brew uninstall gitusr
```

User data files are not automatically removed. To clean them up manually:

```bash
rm -rf "${XDG_DATA_HOME:-$HOME/.local/share}/gitusr"
```

## License

[MIT](./LICENSE)
