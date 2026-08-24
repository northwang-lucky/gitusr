package hook

// GenerateUnifiedZshWrapper returns a single zsh script that contains both
// git() function (clone + commit hooks) and __gitusr_autoload_hook via
// add-zsh-hook chpwd (cd hook). The generated script:
//   - Defines __GITUSR_DATA_DIR for use by hook logic
//   - Defines git() that intercepts clone and commit subcommands
//   - Checks user count first; passes through if only one user is configured
//   - Checks disabled status via "gitusr hooks is-disabled" hidden command
//   - Extracts --gu-name and --gu-email from clone arguments
//   - Applies the configured user from .gitusrrc before commit
//   - Defines __gitusr_autoload_hook() via add-zsh-hook chpwd
//   - Uses zsh syntax (local -a, [[ ]]) and 'command git' to avoid recursion
func GenerateUnifiedZshWrapper() string {
	return `# gitusr unified hook wrapper
__GITUSR_DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/gitusr"

git() {
    # If only one user is configured, pass through directly
    local user_count
    user_count=$(gitusr list 2>/dev/null | wc -l)
    if [[ $user_count -le 1 ]]; then
        command git "$@"
        return $?
    fi

    # Handle clone subcommand
    if [[ "$1" == "clone" ]]; then
        # Check if clone hook is disabled
        if gitusr hooks is-disabled clone 2>/dev/null; then
            command git "$@"
            return $?
        fi
        shift
        local gu_name="" gu_email=""
        local -a remaining_args

        # Extract --gu-name and --gu-email from arguments
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --gu-name=*)
                    gu_name="${1#*=}"
                    shift
                    ;;
                --gu-email=*)
                    gu_email="${1#*=}"
                    shift
                    ;;
                --gu-name)
                    gu_name="$2"
                    shift 2
                    ;;
                --gu-email)
                    gu_email="$2"
                    shift 2
                    ;;
                *)
                    remaining_args+=("$1")
                    shift
                    ;;
            esac
        done

        # Run the actual git clone
        command git clone "${remaining_args[@]}"
        local clone_exit=$?
        if [[ $clone_exit -ne 0 ]]; then
            return $clone_exit
        fi

        # Save the original directory to return later
        local original_dir
        original_dir=$(pwd)

        # Determine target directory and cd into it
        local target_dir=""
        for arg in "${remaining_args[@]}"; do
            case "$arg" in
                -*) ;;
                *)
                    target_dir="$arg"
                    ;;
            esac
        done

        if [[ -n "$target_dir" ]]; then
            local dir_name
            dir_name=$(basename "$target_dir" .git)
            if [[ -d "$dir_name" ]]; then
                \cd "$dir_name" || return 1
            fi
        fi

        # Apply gitusr user identity after clone, by priority:
        #   1. explicit --gu-name/--gu-email args
        #   2. repo-local .gitusrrc, when present
        #   3. host rule from hosts.json, matched against the remote URL
        #   4. interactive gitusr use (legacy behaviour, only without host rules)
        if [[ -n "$gu_name" && -n "$gu_email" ]]; then
            gitusr use --name "$gu_name" --email "$gu_email"
        elif [[ -n "$gu_name" ]]; then
            gitusr use --name "$gu_name"
        elif [[ -n "$gu_email" ]]; then
            gitusr use --email "$gu_email"
        elif [[ -f .gitusrrc ]]; then
            gitusr hooks apply-rc --silent-if-unchanged
        elif [[ -f "$__GITUSR_DATA_DIR/hosts.json" ]]; then
            local clone_url
            clone_url=$(git config --local --get-regexp '^remote\..*\.url$' 2>/dev/null | head -1 | cut -f2)
            if [[ -n "$clone_url" ]]; then
                gitusr hooks apply-host "$clone_url" --silent-if-unchanged
            fi
        else
            gitusr use
        fi

        # Return to the original directory
        \cd "$original_dir" || return 1

        return $clone_exit
    fi

    # Handle commit subcommand
    if [[ "$1" == "commit" ]]; then
        # Check if commit hook is disabled
        if gitusr hooks is-disabled commit 2>/dev/null; then
            command git "$@"
            return $?
        fi
        shift
        if [[ -f .gitusrrc ]]; then
            gitusr hooks apply-rc --silent-if-unchanged
        fi
        command git commit "$@"
        return $?
    fi

    # For all other subcommands, pass through directly
    command git "$@"
}

autoload -U add-zsh-hook

__gitusr_autoload_hook() {
    local gu_count
    gu_count=$(gitusr list 2>/dev/null | wc -l)
    [[ $gu_count -le 1 ]] && return

    # Check if cd hook is disabled
    if gitusr hooks is-disabled cd 2>/dev/null; then
        return
    fi

    if [[ -f .gitusrrc ]]; then
        gitusr hooks apply-rc --silent-if-unchanged
    fi
}

add-zsh-hook -D chpwd __gitusr_autoload_hook
add-zsh-hook chpwd __gitusr_autoload_hook`
}

// Deprecated: GenerateZshWrapper is the old git-only wrapper.
// Use GenerateUnifiedZshWrapper instead, which combines git() and
// __gitusr_autoload_hook (chpwd) into a single script with disabled
// hook checks.
//
// GenerateZshWrapper returns a zsh shell wrapper script that wraps git
// commands with gitusr integration. The generated script:
//   - Defines a git() function that intercepts clone and commit subcommands
//   - Checks user count first and passes through if only one user is configured
//   - Extracts --gu-name and --gu-email from clone arguments
//   - Applies the configured user from .gitusrrc before commit
//   - Uses 'command git' to avoid recursion and '\cd' to avoid alias expansion
func GenerateZshWrapper() string {
	return `git() {
    # If only one user is configured, pass through directly
    local user_count
    user_count=$(gitusr list 2>/dev/null | wc -l)
    if [ "$user_count" -le 1 ]; then
        command git "$@"
        return $?
    fi

    # Handle clone subcommand
    if [ "$1" = "clone" ]; then
        shift
        local gu_name="" gu_email=""
        local -a remaining_args

        # Extract --gu-name and --gu-email from arguments
        while [ $# -gt 0 ]; do
            case "$1" in
                --gu-name=*)
                    gu_name="${1#*=}"
                    shift
                    ;;
                --gu-email=*)
                    gu_email="${1#*=}"
                    shift
                    ;;
                --gu-name)
                    gu_name="$2"
                    shift 2
                    ;;
                --gu-email)
                    gu_email="$2"
                    shift 2
                    ;;
                *)
                    remaining_args+=("$1")
                    shift
                    ;;
            esac
        done

        # Run the actual git clone
        command git clone "${remaining_args[@]}"
        local clone_exit=$?
        if [ $clone_exit -ne 0 ]; then
            return $clone_exit
        fi

        # Save the original directory to return later
        local original_dir
        original_dir=$(pwd)

        # Determine target directory and cd into it
        local target_dir=""
        for arg in "${remaining_args[@]}"; do
            case "$arg" in
                -*) ;;
                *)
                    target_dir="$arg"
                    ;;
            esac
        done

        if [ -n "$target_dir" ]; then
            local dir_name
            dir_name=$(basename "$target_dir" .git)
            if [ -d "$dir_name" ]; then
                \cd "$dir_name" || return 1
            fi
        fi

        # Always apply gitusr user identity after clone
        if [ -n "$gu_name" ] && [ -n "$gu_email" ]; then
            gitusr use --name "$gu_name" --email "$gu_email"
        elif [ -n "$gu_name" ]; then
            gitusr use --name "$gu_name"
        elif [ -n "$gu_email" ]; then
            gitusr use --email "$gu_email"
        else
            gitusr use
        fi

        # Return to the original directory
        \cd "$original_dir" || return 1

        return $clone_exit
    fi

    # Handle commit subcommand
    if [ "$1" = "commit" ]; then
        shift
        if [ -f ".gitusrrc" ]; then
            gitusr hooks apply-rc --silent-if-unchanged
        fi
        command git commit "$@"
        return $?
    fi

    # For all other subcommands, pass through directly
    command git "$@"
}`
}
