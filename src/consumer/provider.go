package consumer

import "time"

//var testp protocol.CustRequest

type provider struct {
	name            string
	info            providerInfo
	weight          uint32
	consumer        *Consumer
	isFull          bool
	fullLevel       int
	baseDelay       int64
	baseDelaySample int
	chanDelay       chan time.Duration

	//connections []connection
	//chanOut     chan protocol.CustRequest
	//chanIn      chan protocol.CustResponse
}

func (p *provider) maintain() {
	for {
		select {
		case d := <-p.chanDelay:
			if p.baseDelaySample < baseDelaySampleSize {
				p.baseDelay = (p.baseDelay + d.Nanoseconds()) / 2
				p.baseDelaySample ++
			} else {
				if !p.isFull && d.Nanoseconds() > int64(float64(p.baseDelay)*float64(delayTimes)) {
					p.fullLevel ++
					if p.fullLevel > fullMaxLevel {
						p.isFull = true
						return
					}
				} else {
					p.fullLevel = 0
				}
			}
		}
	}
}

//func (p *provider) tryConnect() {
//	p.connections = make([]connection, 16)
//	testp = protocol.CustRequest{
//		Identifier: 1,
//		Content:    []byte("rejected"),
//	}
//	//ERROR Handling
//	p.pressureTestEach()
//	for {
//
//	}
//
//}
//
//func (p *provider) pressureTestEach() {
//	p.chanOut <- testp
//}
//
//func (p *provider) pressureTest() {
//	for {
//
//	}
//
//}

type providerInfo struct {
	IP     string
	Weight uint32
}
