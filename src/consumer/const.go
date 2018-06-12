package consumer

import (
	"time"
)

const (
	logger              = true
	dialTimeout         = 5 * time.Second
	queueSize           = 128
	overLoadSize		= 64
	headerMaxSize       = 4
	delayTimes          = 1.25
	fullMaxLevel        = 10
	baseDelaySampleSize = 100
	connMinSize         = 16
	connMaxSize         = 1024
	listenPort          = "20000"
	requestPort         = "30000"
	//simulateQPS   = 2000
	//bodyMaxSize   = 4096
)
