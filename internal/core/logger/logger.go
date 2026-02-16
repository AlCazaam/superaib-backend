package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel) // Default level
}

// SetLevel sets the logging level dynamically
func SetLevel(level logrus.Level) {
	Log.SetLevel(level)
}
