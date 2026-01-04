package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"dummy-chain/metadata"

	"github.com/urfave/cli/v2"
)

var (
	app = cli.NewApp()
	m   *Manager
)

func Run() {
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Stop() {
	err := m.Stop()
	if err != nil {
		panic(err)
	}
}

func init() {
	app.Name = filepath.Base(os.Args[0])
	app.Usage = "Node"
	app.HideVersion = true
	app.HideHelpCommand = true
	app.Version = metadata.Version
	app.Compiled = time.Now()
	app.Commands = []*cli.Command{
		versionCommand,
		sendCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Before = beforeAction
	app.Action = action
	app.After = afterAction
}

func beforeAction(ctx *cli.Context) error {
	maxCpu := runtime.NumCPU()

	fmt.Printf(`Node - Runtime Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Current time:	%v
Version:	%s
Git commit:	%s
Build Date:	%s
Max CPU cores:	%d
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, time.Now().UTC().Format("2006-01-02 15:04:05"), metadata.Version, metadata.GitCommit, metadata.BuildDate, maxCpu)
	runtime.GOMAXPROCS(maxCpu)

	return nil
}
func action(ctx *cli.Context) error {
	var err error
	m, err = NewManager(ctx)
	if err != nil {
		return err
	}

	return m.Start()

}
func afterAction(*cli.Context) error {
	return nil
}
