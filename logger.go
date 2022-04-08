package pirsch

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[pirsch] ", log.LstdFlags)
