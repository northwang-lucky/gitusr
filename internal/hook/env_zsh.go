package hook

// GenerateZshEnv returns a zsh script for chpwd-based .gitusrrc detection.
// The generated script:
//   - Uses autoload -U add-zsh-hook (zsh native)
//   - Defines __gitusr_autoload_hook() that checks user count and applies .gitusrrc
//   - Uses add-zsh-hook -D first to remove old hook (idempotency)
//   - Then add-zsh-hook chpwd to register
func GenerateZshEnv() string {
	return `autoload -U add-zsh-hook

__gitusr_autoload_hook() {
    local gu_count
    gu_count=$(gitusr list 2>/dev/null | wc -l)
    [[ $gu_count -le 1 ]] && return

    if [[ -f .gitusrrc ]]; then
        gitusr hook apply-rc --silent-if-unchanged
    fi
}

add-zsh-hook -D chpwd __gitusr_autoload_hook
add-zsh-hook chpwd __gitusr_autoload_hook`
}
