package hook

// HookType represents the type of hook (clone or commit)
type HookType string

const (
	HookTypeClone  HookType = "clone"
	HookTypeCommit HookType = "commit"
	HookTypeCD     HookType = "cd"
)

// AllHookTypes lists all hook types in their canonical iteration order.
var AllHookTypes = []HookType{HookTypeClone, HookTypeCommit, HookTypeCD}

// ShellType represents the shell type (bash or zsh)
type ShellType string

const (
	ShellTypeBash ShellType = "bash"
	ShellTypeZsh  ShellType = "zsh"
)

// HookState tracks which hooks are installed and which are disabled
type HookState struct {
	InstalledTypes []HookType `json:"installed_types"`
	DisabledTypes  []HookType `json:"disabled_types"`
}

// defaultShells returns the default list of supported shells (bash and zsh).
func defaultShells() []ShellType {
	return []ShellType{ShellTypeBash, ShellTypeZsh}
}

// GitusrRC represents the .gitusrrc file structure
type GitusrRC struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// HookInstallResult represents the result of installing a hook
type HookInstallResult struct {
	Type     HookType
	Shell    ShellType
	FilePath string
}
