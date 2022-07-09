package util

import (
	"log"
	"os"
)

// GetDefaultLogger returns the default logger.
func GetDefaultLogger() *log.Logger {
	return log.New(os.Stdout, "[pirsch] ", log.LstdFlags)
}
