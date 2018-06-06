package consumer

import (
	"time"
)

const (
	logger      = false
	dialTimeout = 5 * time.Second
	queueSize   = 1024
	connsSize   = 256
)

type ProviderInfo struct {
	IP     string
	Weight uint32
}
