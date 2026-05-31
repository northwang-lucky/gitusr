package main

import (
	"os"
	"path/filepath"

	"gitusr/internal/cli"
	"gitusr/internal/format"
	"gitusr/internal/i18n"
	"gitusr/internal/store"
	"gitusr/internal/xdgpath"
)

func main() {
	path, err := xdgpath.DataFilePath()
	if err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}

	i18n.Init()

	s := store.NewJSONStore(path)

	cmdName := filepath.Base(os.Args[0])
	rootCmd := cli.NewRootCmd(s, cmdName)

	if err := rootCmd.Execute(); err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}
}
