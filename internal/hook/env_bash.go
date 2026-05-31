package hook

// GenerateBashEnv returns a bash script for cd-based .gitusrrc detection.
// The generated script:
//   - Defines __gitusr_use_if_found() that checks user count and applies .gitusrrc
//   - Defines __gitusrcd() that wraps cd and triggers __gitusr_use_if_found
//   - Uses \cd to avoid recursive alias expansion
//   - Aliases cd to __gitusrcd
func GenerateBashEnv() string {
	return `__gitusr_use_if_found() {
    local gu_count
    gu_count=$(gitusr hook count-users 2>/dev/null || echo 0)
    [ "$gu_count" -le 1 ] && return

    if [ -f .gitusrrc ]; then
        gitusr hook apply-rc --silent-if-unchanged
    fi
}

__gitusrcd() {
    \cd "$@" || return $?
    __gitusr_use_if_found
}

alias cd=__gitusrcd`
}
