package conf

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// LogConfig specifies all the parameters needed for logging
type LogConfig struct {
	Level string
}

// ConfigureLogging will take the logging configuration and also adds
// a few default parameters
func ConfigureLogging(config *LogConfig) error {
	level, err := logrus.ParseLevel(strings.ToUpper(config.Level))
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	// always use the fulltimestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	return nil
}
