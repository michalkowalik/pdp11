package unibus

import (
	"fmt"
	"log"
	"os"
)

// PLogger wraps stdlib logger functionality
type PLogger struct {
	logger *log.Logger
}

func initLogger(path string) *PLogger {
	logger := new(PLogger)
	if len(path) < 1 {
		fmt.Printf("logging to stdout")
		logger.logger = log.New(os.Stdout, "pdp", log.Ldate|log.Ltime)
	} else {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("can't create debug logger!")
		}
		defer file.Close()
		logger.logger = log.New(file, "pdp", log.Ltime|log.Ldate)
	}
	return logger
}

func (l *PLogger) info(msg string) {
	l.logger.Printf("INFO: %s\n", msg)
}
