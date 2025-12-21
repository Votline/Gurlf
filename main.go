package main

import (
	"flag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogger(d *bool) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.DisableStacktrace = true
	
	lvl := zapcore.ErrorLevel
	if *d {lvl = zapcore.DebugLevel}
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
	f := args[0]

	log.Debug("file", zap.String("f", f))

}
