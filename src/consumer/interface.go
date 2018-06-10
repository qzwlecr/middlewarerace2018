package consumer

import (
	"time"
)

const (
	logger          = true
	dialTimeout     = 5 * time.Second
	checkTimeout    = 2 * time.Second
	queueSize       = 1024
	connsSize       = 256
	oldWeight       = 0
	newWeight       = 10 - oldWeight
	baseTime        = 50000
	decreaseTimeout = 200 * time.Millisecond
	decreaseTime    = 1000 * 1000
)

type ProviderInfo struct {
	IP     string
	Weight uint32
}
