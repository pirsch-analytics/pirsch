package db

import (
	"math/rand"
	"time"
)

func init() {
	// TODO
	rand.Seed(time.Now().UnixNano())
}
