package consumer

import (
	"time"
)

const (
	logger        = false
	dialTimeout   = 5 * time.Second
	queueSize     = 512
	headerMaxSize = 4
	ignoreSize    = 100
	listenPort    = "20000"
	requestPort   = "30000"
	//overLoadSize        = 5
	//simulateQPS   = 2000
	//bodyMaxSize   = 4096
	//delayTimes          = 1.25
	//fullMaxLevel        = 10
	//baseDelaySampleSize = 100
	//connMinSize         = 16
	//connMaxSize         = 768
)
