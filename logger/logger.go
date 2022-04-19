package logger

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const CmdLogLevel = "log-level"
const CmdLogFormat = "log-format"
const CmdLogOutputs = "log-outputs"

const DefaultLogLevel = "info"
const DefaultLogFormat = "console"

var DefaultLogOutputs = []string{"stderr"}

var L *zap.SugaredLogger

func ConfigCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String(CmdLogLevel, DefaultLogLevel, "logging level (debug | info | warn | error | dpanic | panic | fatal)")
	cmd.PersistentFlags().String(CmdLogFormat, DefaultLogFormat, "logging format (json | console)")
	cmd.PersistentFlags().StringArray(CmdLogOutputs, DefaultLogOutputs, "logging outputs (stdout | stderr | /somewhere/to/some/file)")
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
	cmdOutputs, err := cmd.Flags().GetStringArray(CmdLogOutputs)
	if err != nil {
		panic(err)
	}
	var level zapcore.Level
	err = level.UnmarshalText([]byte(cmdLevel))
	if err != nil {
		fmt.Println("Unable to decode log level, using default level INFO")
		level = zapcore.InfoLevel
	}
	SetupLogger(level, cmdOutputs, cmdFormat)
}

func SetupLogger(level zapcore.Level, cmdOutputs []string, cmdFormat string) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.OutputPaths = cmdOutputs
	cfg.ErrorOutputPaths = cmdOutputs
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
