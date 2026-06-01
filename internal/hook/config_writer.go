package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/northwang-lucky/gitusr/internal/xdgpath"
)

const (
	markerBegin    = "# gitusr hook begin"
	markerEnd      = "# gitusr hook end"
	markerCDBegin  = "# gitusr cd begin"
	markerCDEnd    = "# gitusr cd end"
)

// appendSourceLine is a compatibility wrapper retained for existing tests.
// Deprecated: use AppendSourceLine instead.
func appendSourceLine(configPath, wrapperPath string) error {
	block := fmt.Sprintf("%s\n%s\n%s\n", markerBegin, "source "+wrapperPath, markerEnd)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read config: %w", err)
		}
		data = []byte{}
	}
	content := string(data)
	content = removeMarkedBlock(content)
	content = strings.TrimRight(content, "\n")
	if content != "" {
		content += "\n"
	}
	content += block
	return os.WriteFile(configPath, []byte(content), 0644)
}

// WriteWrapperFile writes a shell wrapper script to the hooks directory.
// The path is: {XDG_DATA_HOME}/gitusr/hooks/git-wrapper.{sh|zsh}.
// The directory is created if it doesn't exist (0755 permissions).
// The file is written with 0644 permissions.
// Returns the absolute path to the written file.
func WriteWrapperFile(shell ShellType, content string) (string, error) {
	ext, err := wrapperExt(shell)
	if err != nil {
		return "", err
	}

	dir, err := hooksDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create hooks dir: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("git-wrapper.%s", ext))
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write wrapper file: %w", err)
	}

	return path, nil
}

// AppendSourceLine appends a marked source block to the shell config file.
// Shell config: ~/.bashrc for bash, ~/.zshrc for zsh.
// The block is:
//
//	# gitusr hook begin
//	source /path/to/wrapper
//	# gitusr hook end
//
// If the markers already exist, the old block is removed first (idempotent).
func AppendSourceLine(shell ShellType, wrapperPath string) error {
	configPath, err := shellConfigPath(shell)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read shell config: %w", err)
		}
		data = []byte{}
	}

	content := string(data)
	content = removeMarkedBlock(content)
	content = strings.TrimRight(content, "\n")

	block := fmt.Sprintf("\n# gitusr hook begin\nsource %s\n# gitusr hook end\n", wrapperPath)
	content += block

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write shell config: %w", err)
	}

	return nil
}

// RemoveSourceBlock removes the marked source block from the shell config file.
// If the block is not found, it returns nil (no error).
func RemoveSourceBlock(shell ShellType) error {
	configPath, err := shellConfigPath(shell)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read shell config: %w", err)
	}

	content := string(data)
	newContent := removeMarkedBlock(content)
	if newContent == content {
		return nil
	}

	return os.WriteFile(configPath, []byte(newContent), 0644)
}

// wrapperExt returns the file extension for the given shell type.
func wrapperExt(shell ShellType) (string, error) {
	switch shell {
	case ShellTypeBash:
		return "sh", nil
	case ShellTypeZsh:
		return "zsh", nil
	default:
		return "", fmt.Errorf("unsupported shell type: %s", shell)
	}
}

func wrapperFileName(hookType HookType, ext string) string {
	if hookType == HookTypeCD {
		return fmt.Sprintf("cd-env.%s", ext)
	}
	return fmt.Sprintf("git-wrapper.%s", ext)
}

// hooksDir returns the hooks directory path under the XDG data directory.
func hooksDir() (string, error) {
	userListPath, err := xdgpath.DataFilePath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(userListPath)
	return filepath.Join(dir, "hooks"), nil
}

// shellConfigPath returns the path to the shell config file
// based on the current user's home directory.
func shellConfigPath(shell ShellType) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch shell {
	case ShellTypeBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellTypeZsh:
		return filepath.Join(home, ".zshrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell type: %s", shell)
	}
}

// removeMarkedBlock removes the content between and including the marker lines.
// Markers: "# gitusr hook begin" and "# gitusr hook end".
// If either marker is not found, the original content is returned unchanged.
func removeMarkedBlock(content string) string {
	startMarker := "# gitusr hook begin"
	endMarker := "# gitusr hook end"

	lines := strings.Split(content, "\n")
	startLine := -1
	endLine := -1

	for i, line := range lines {
		if strings.Contains(line, startMarker) {
			startLine = i
		}
		if strings.Contains(line, endMarker) {
			endLine = i
			break
		}
	}

	if startLine == -1 || endLine == -1 {
		return content
	}

	// Rebuild content without the block
	var result strings.Builder
	for i, line := range lines {
		if i < startLine {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(line)
		} else if i > endLine {
			result.WriteString("\n")
			result.WriteString(line)
		}
	}

	return result.String()
}
