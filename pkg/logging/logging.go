package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Logger is the logger to use in azp-agent-autoscaler
var Logger = log.Logger{
	Out: os.Stderr,
	Formatter: &log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	},
	Hooks:        make(log.LevelHooks),
	Level:        log.InfoLevel,
	ExitFunc:     os.Exit,
	ReportCaller: false,
}
