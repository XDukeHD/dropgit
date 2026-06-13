package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(level string) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Log.SetOutput(os.Stdout)
	
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		Log.SetLevel(logrus.InfoLevel)
		Log.Warnf("Invalid log level '%s', defaulting to info", level)
	} else {
		Log.SetLevel(lvl)
	}
}
