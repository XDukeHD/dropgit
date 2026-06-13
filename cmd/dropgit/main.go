package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/XDukeHD/dropgit/internal/daemon"
)

func main() {
	runOnce := flag.Bool("once", false, "Run backup once and exit")
	version := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *version {
		fmt.Println("DropGit version 1.0.0")
		os.Exit(0)
	}

	if *runOnce {
		daemon.RunOnce()
	} else {
		daemon.Run()
	}
}
