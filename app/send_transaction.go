package app

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	sendCommand = &cli.Command{
		Action:    sendAction,
		Name:      "send",
		Usage:     "Send a transaction",
		ArgsUsage: "0xaddress amount[,integer]",
	}
)

func sendAction(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return errors.New("invalid arguments")
	}
	to := c.Args().Get(0)
	value := c.Args().Get(1)

	var err error
	m, err = NewManager(c)
	if err != nil {
		return err
	}
	if err = m.node.Start(false); err != nil {
		m.logger.Fatal("failed to start node", zap.String("reason", err.Error()))
		os.Exit(1)
	}

	return m.node.SendTransaction(to, value)
}
