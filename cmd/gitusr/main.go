package main

import (
	"os"
	"path/filepath"

	"github.com/northwang-lucky/gitusr/internal/cli"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	"github.com/northwang-lucky/gitusr/internal/store"
	"github.com/northwang-lucky/gitusr/internal/xdgpath"
)

func main() {
	userPath, err := xdgpath.DataFilePath()
	if err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}

	hostPath, err := xdgpath.HostsFilePath()
	if err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}

	i18n.Init()

	s := store.NewJSONStore(userPath)
	hostStore := store.NewJSONHostRuleStore(hostPath)

	cmdName := filepath.Base(os.Args[0])
	rootCmd := cli.NewRootCmd(s, hostStore, cmdName)

	if err := rootCmd.Execute(); err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}
}
