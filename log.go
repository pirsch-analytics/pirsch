package pirsch

import (
	"log"
	"os"
)

// logger is the global default logger if none is provided in configuration.
var logger = log.New(os.Stdout, logPrefix, log.LstdFlags)
