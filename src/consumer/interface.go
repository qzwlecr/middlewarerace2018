package consumer

import (
	"time"
)

const (
	logger       = false
	dialTimeout  = 5 * time.Second
	checkTimeout = 2 * time.Second
	queueSize    = 1024
	connsSize    = 256
	oldWeight    = 0
	newWeight    = 10 - oldWeight
	baseTime = 50000
)

type ProviderInfo struct {
	IP     string
	Weight uint32
}
