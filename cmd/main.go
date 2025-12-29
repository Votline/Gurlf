package main

import (
	"flag"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Votline/Gurlf"
)

func initLogger(d *bool) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.ConsoleSeparator = " | "

	lvl := zapcore.ErrorLevel
	if *d {
		lvl = zapcore.DebugLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	log, _ := cfg.Build()

	return log
}

func main() {
	d := flag.Bool("debug", false, "use for debug mode")
	flag.Parse()

	log := initLogger(d)

	args := flag.Args()
	if len(args) < 1 {
		log.Error("Specify the path to file")
		return
	}
	p := args[0]
	
	data, err := gurlf.Scan(p)
	if err != nil {
		log.Error("Scan failed", zap.Error(err))
	}

	s := struct{
		ID int `gurlf:"ID"`
		Body string `gurlf:"BODY"`
		Hdrs string `gurlf:"HEADERS"`
	}{}

	err = gurlf.Unmarshal(data[1], &s)
	if err != nil {
		log.Error("Unmarshal failed", zap.Error(err))
	}

	fmt.Printf("\n%s\n%s\n%s\n", s.ID, s.Body, s.Hdrs)
}
