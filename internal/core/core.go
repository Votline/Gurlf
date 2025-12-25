package core

import (
	"fmt"
	"gurlf/internal/scanner"

	"go.uber.org/zap"
)

type core struct {
	fileName string
	log      *zap.Logger
}

func New(f string, l *zap.Logger) *core {
	return &core{fileName: f, log: l}
}

func (c *core) Start() error {
	const op = "core.Start"

	c.log.Debug("starting process",
		zap.String("file", c.fileName))

	data, err := scanner.Scan(c.fileName, c.log)
	if err != nil {
		return fmt.Errorf("%s: scan file: %w", op, err)
	}

	c.log.Debug("scan complete",
		zap.Int("configs length", len(data)))

	return nil
}
