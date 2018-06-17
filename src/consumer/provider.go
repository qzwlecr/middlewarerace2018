package consumer

import (
	"protocol"
	"sync"
	"time"
)

//var testp protocol.CustRequest

type provider struct {
	name        string
	info        providerInfo
	weight      uint32
	consumer    *Consumer
	activeMu    sync.Mutex
	chanRequest chan protocol.CustRequest
	chanDelay   chan time.Duration
	delay       int64
	active      int
	capacity    int
}

func (p *provider) updateDelay() {
	for {
		select {
		case d := <-p.chanDelay:
			p.delay = (p.delay + d.Nanoseconds()) / 2
		case _ = <-time.After(500 * time.Millisecond):
			p.delay = 0
		}
	}
}

//func (p *provider) maintain() {
//	for {
//		select {
//		case d := <-p.chanDelay:
//			if p.baseDelaySample < baseDelaySampleSize {
//				log.Println("Provider", p.info, " with base delay: ", p.baseDelay)
//				p.baseDelay = (p.baseDelay + d.Nanoseconds()) / 2
//				p.baseDelaySample ++
//			} else {
//				if !p.isFull && d.Nanoseconds() > int64(float64(p.baseDelay)*float64(delayTimes)) {
//					log.Println(p.info, "comes to full:", d.Nanoseconds())
//					p.fullLevel ++
//					if p.fullLevel > fullMaxLevel {
//						p.isFull = true
//						log.Println(p.info, "is full.")
//						return
//					}
//				} else {
//					p.fullLevel --
//				}
//			}
//		}
//	}
//}

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
//	p.chanRequest <- testp
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
