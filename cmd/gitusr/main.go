package main

import (
	"os"

	"gitusr/internal/cli"
	"gitusr/internal/format"
	"gitusr/internal/store"
	"gitusr/internal/xdgpath"
)

func main() {
	path, err := xdgpath.DataFilePath()
	if err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}

	s := store.NewJSONStore(path)
	rootCmd := cli.NewRootCmd(s)

	if err := rootCmd.Execute(); err != nil {
		format.PrintErr(err.Error())
		os.Exit(1)
	}
}
