package consumer

import (
	"time"
)

const (
	logger              = true
	dialTimeout         = 5 * time.Second
	queueSize           = 512
	overLoadSize        = 10
	headerMaxSize       = 4
	bodyMaxSize   = 4096
	delayTimes          = 1.25
	fullMaxLevel        = 10
	baseDelaySampleSize = 100
	connMinSize         = 16
	connMaxSize         = 1024
	ignoreSize          = 50
	listenPort          = "20000"
	requestPort         = "30000"
	//simulateQPS   = 2000
)
