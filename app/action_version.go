package app

import (
	"dummy-chain/metadata"
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	versionCommand = &cli.Command{
		Action:    versionAction,
		Name:      "version",
		Usage:     "Print version numbers",
		ArgsUsage: " ",
	}
)

func versionAction(c *cli.Context) error {
	fmt.Printf(`Node - Version Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Version:        %s
Role:           %s
Git Commit:     %s
Build Date:     %s
Go Version:     %s
Architecture:   %s
OS:             %s
Compiler:       %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, metadata.Version, metadata.Role, metadata.GitCommit, metadata.BuildDate, runtime.Version(), runtime.GOARCH, runtime.GOOS, runtime.Compiler)

	return nil
}
