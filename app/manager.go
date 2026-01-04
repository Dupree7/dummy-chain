package app

import (
	"dummy-chain/common"
	"dummy-chain/common/config"
	"dummy-chain/node"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type Manager struct {
	ctx    *cli.Context
	node   *node.Node
	logger *zap.Logger
}

func NewManager(ctx *cli.Context) (*Manager, error) {
	newConfig, err := MakeConfig()
	if err != nil {
		return nil, err
	}

	logger, err := common.CreateLogger()
	if err != nil {
		return nil, err
	}

	newNode, err := node.NewNode(newConfig, logger)

	if err != nil {
		logger.Info("failed to create the node", zap.String("reason", err.Error()))
		return nil, err
	}

	return &Manager{
		ctx:    ctx,
		node:   newNode,
		logger: logger,
	}, nil
}

func (m *Manager) Start() error {
	// Start up the node
	m.logger.Info("Preparing ...")
	if err := m.node.Start(true); err != nil {
		m.logger.Fatal("failed to start node", zap.String("reason", err.Error()))
		os.Exit(1)
	} else {
		m.logger.Info("Ready ... ")
	}

	signalFromOutside := false
	// Listening event closes the node
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		defer signal.Stop(c)
		<-c
		signalFromOutside = true
		m.logger.Info("Shutting down node from go func")

		go func() {
			if err := m.Stop(); err != nil {
				m.logger.Error(err.Error())
			}
		}()

		for i := 10; i > 0; i-- {
			<-c
			if i > 1 {
				m.logger.Warn("Please DO NOT interrupt the shutdown process, panic may occur.", zap.String("times", string(rune(i-1))))
			}
		}
	}()

	// Waiting for node to close
	m.node.Wait()
	if signalFromOutside == false {
		if err := m.Stop(); err != nil {
			m.logger.Info(err.Error())
		}
	}

	return nil
}

func (m *Manager) Stop() error {
	m.logger.Warn("Stopping node ...")

	//if err := m.SaveConfig(); err != nil {
	//	m.logger.Info("Failed to save config", zap.String("reason", err.Error()))
	//}

	if err := m.node.Stop(); err != nil {
		m.logger.Info("Failed to stop node", zap.String("reason", err.Error()))
	} else {
		m.logger.Info("successfully stopped node")
	}
	return nil
}

func (m *Manager) SaveConfig() error {
	m.logger.Info("Write config to file")
	conf := m.node.GetGlobalConfig()
	if conf != nil {
		m.logger.Info("wrote config at the end")
		return config.WriteConfig(conf)
	}
	return errors.New("Invalid config")
}
