package logger

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const CmdLogLevel = "log-level"
const CmdLogFormat = "log-format"

const DefaultLogLevel = "info"
const DefaultLogFormat = "console"

var L *zap.SugaredLogger

func ConfigCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String(CmdLogLevel, DefaultLogLevel, "logging level")
	cmd.PersistentFlags().String(CmdLogFormat, DefaultLogFormat, "logging format (json | console)")
}

func SetupLoggerFromCmdArgs(cmd *cobra.Command) {
	cmdLevel, err := cmd.Flags().GetString(CmdLogLevel)
	if err != nil {
		panic(err)
	}
	cmdFormat, err := cmd.Flags().GetString(CmdLogFormat)
	if err != nil {
		panic(err)
	}
	var level zapcore.Level
	err = level.UnmarshalText([]byte(cmdLevel))
	if err != nil {
		fmt.Println("Unable to decode log level, using default level INFO")
		level = zapcore.InfoLevel
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.Encoding = cmdFormat
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	L = l.Sugar()
}
