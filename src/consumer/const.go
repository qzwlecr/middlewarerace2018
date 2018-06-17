package consumer

import (
	"time"
)

const (
	logger              = true
	dialTimeout         = 5 * time.Second
	queueSize           = 512
	overLoadSize        = 5
	headerMaxSize       = 4
	delayTimes          = 1.25
	fullMaxLevel        = 10
	baseDelaySampleSize = 100
	connMinSize         = 16
	connMaxSize         = 768
	ignoreSize          = 100
	listenPort          = "20000"
	requestPort         = "30000"
	//simulateQPS   = 2000
	//bodyMaxSize   = 4096
)
