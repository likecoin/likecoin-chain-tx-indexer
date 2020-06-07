package logger

import "go.uber.org/zap"

var L *zap.SugaredLogger

func init() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	L = l.Sugar()
}
