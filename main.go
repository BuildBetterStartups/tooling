package main

import (
	"github.com/navenprasad/tooling/commander"
	"github.com/navenprasad/tooling/dirscanner"
)

func main() {
    dirscanner.ScanAndCleanGitHubDirs()
    commander.RunCommander()
}