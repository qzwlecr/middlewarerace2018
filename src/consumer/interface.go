package consumer

import (
	"time"
)

const (
	logger       = true
	dialTimeout  = 5 * time.Second
	checkTimeout = 2 * time.Second
	queueSize    = 1024
	connsSize    = 512
	oldWeight    = 8
	newWeight    = 10 - oldWeight
)

type ProviderInfo struct {
	IP     string
	Weight uint32
}
