package provider

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"protocol"
	"strconv"
	"time"
	"utility/rrselector"
	"utility/timing"

	etcdv3 "github.com/coreos/etcd/clientv3"
)

// change it to your listen address and provider address
const (
	lnAddr       = ":30000"
	providerAddr = ":20880"
	targetConns  = 3
)

//NewProvider receive etcd server address, the service name, and the service info.
//And return the consumer which has been started.
func NewProvider(endpoints []string, name string, info ProviderInfo) *Provider {

	cfg := etcdv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	cli, err := etcdv3.New(cfg)

	if err != nil {
		log.Fatal(err)
	}

	p := &Provider{
		name:     name,
		etcdAddr: endpoints,
		info:     info,
		chanStop: make(chan error),
		client:   cli,
		connIn:   make([]chan []byte, targetConns),
		connOut:  make([]chan []byte, targetConns),
	}

	// then, pre-create these connections..

	go p.Start()

	return p

}

//Start shouldn't be called manually.
func (p *Provider) Start() {
	ch := p.keepAlive()

	ln, err := net.Listen("tcp", lnAddr)
	if err != nil {
		p.revoke()
		log.Fatal(err)
	}

	tcpCh := make(chan int)

	var converter protocol.SimpleConverter
	// handler for listening over tcp
	go p.handleReq(ln, tcpCh, &converter)

	go func(ch <-chan *etcdv3.LeaseKeepAliveResponse, tcpCh chan<- int) {
		// close the tcp listener
		defer func() {
			tcpCh <- 0
		}()

		for {
			select {
			case <-p.chanStop:
				p.revoke()
				return
			case _, ok := <-ch:
				if !ok {
					p.revoke()
					return
				}
			case <-p.client.Ctx().Done():
				return
			}
		}
	}(ch, tcpCh)
}

// maintaing timing map
type tMapEntry struct {
	id   [8]byte
	tBeg time.Time
}

func (p *Provider) handleReq(ln net.Listener, tcpCh <-chan int, converter *protocol.SimpleConverter) {
	defer ln.Close()
	var connSelector rrselector.RRSelector
	go func(converter *protocol.SimpleConverter) {

		for {
			tm := time.Now()
			cConn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}

			// cConn.Close()

			// now we will try to pre-create some fixed amount
			// of provider connections!
			// -- the pre-create behavior seems not working well: provider starts too slow!
			// let's try something else..
			for p.createdConn < targetConns {
				pConn, err := net.Dial("tcp", providerAddr)
				if err != nil {
					panic("Unable to connect to provider while creating No." + strconv.Itoa(p.createdConn) + " connections. Err is " + err.Error())
				}
				pReqMsg := make(chan []byte, 100)
				pRespMsg := make(chan []byte, 100)
				go providerWrite(pConn, pReqMsg)
				go providerRead(pConn, pRespMsg)
				p.connIn[p.createdConn] = pReqMsg
				p.connOut[p.createdConn] = pRespMsg
				p.createdConn++
			}
			/*pConn, err := net.Dial("tcp", providerAddr)
			if err != nil {
				log.Fatal(err)
			}
			// pConn.Close()
			*/
			rrvalue := connSelector.SelectBetween(targetConns)
			// from client read
			cReqMsg := make(chan []byte, 10)
			// instead of create new instances.. now we will try some brand new technologies!
			// pReqMsg := make(chan []byte, 10)
			// pRespMsg := make(chan []byte, 10)
			pReqMsg := p.connIn[rrvalue]
			pRespMsg := p.connOut[rrvalue]
			cRespMsg := make(chan []byte, 10)
			go clientRead(cConn, cReqMsg)

			// Timing part. Will not be used anymore!
			/*
				elapsedCh := make(chan int64, 10)
				addCh := make(chan tMapEntry, 5)
				delCh := make(chan [8]byte, 5)
				getReqCh := make(chan [8]byte, 1)
				getRetCh := make(chan time.Time, 1)
				go convertRequest(addCh, delCh, getReqCh, getRetCh)
			*/

			// to provider converter
			go tpConvert(converter, cReqMsg, pReqMsg)

			// to server write - Commented out, see ABOVE!
			// go providerWrite(pConn, pReqMsg)

			// from server read - Commented out, see ABOVE!
			// go providerRead(pConn, pRespMsg)

			// from provider converter
			go tcConvert(converter, pRespMsg, cRespMsg)

			// to client write
			go clientWrite(cConn, cRespMsg)
			timing.Since(tm, "HAND Provider//HandleReq < EACH")
		}
	}(converter)

	<-tcpCh
}

func clientRead(cConn net.Conn, cReqMsg chan<- []byte) {
	defer cConn.Close()
	for {
		tm := time.Now()
		bl := make([]byte, 4)
		_, err := io.ReadFull(cConn, bl)
		if err != nil {
			log.Println("failed to read length", err)
			return
		}

		lens := binary.BigEndian.Uint32(bl)

		cbreq := make([]byte, lens)
		_, err = io.ReadFull(cConn, cbreq)
		if err != nil {
			log.Println("failed to read content", err)
			return
		}

		// the mini-pressurer stuff
		tmagic := binary.BigEndian.Uint64(cbreq[8:16])
		if tmagic == protocol.CUST_MAGIC {
			// this is the mini-pressurer little thing!
			go func(cbreq []byte) {
				<-time.After(50 * time.Millisecond)
				// DIRECT RETURN
				bl := make([]byte, 4)
				direp := protocol.CustResponse{
					Identifier: binary.BigEndian.Uint64(cbreq[:8]),
					Delay:      protocol.CUST_MAGIC,
					Reply:      make([]byte, 1),
				}
				cbrep, _ := direp.ToByteArr()
				binary.BigEndian.PutUint32(bl, uint32(len(cbrep)))

				_, err := cConn.Write(bl)
				if err != nil {
					log.Println(err)
					return
				}

				//log.Println("to customer", cbrep)
				_, err = cConn.Write(cbrep)
				if err != nil {
					log.Println(err)
					return
				}
			}(cbreq)
			continue
		}

		//log.Println("msg to cReqMsg", cbreq)
		cReqMsg <- cbreq
		timing.Since(tm, "READ Provider//clientRead < EACH Req")
	}
}

// Do you need more timing?
/*
func convertRequest(addCh <-chan tMapEntry, delCh, getReqCh <-chan [8]byte, getRetCh chan<- time.Time) {
	tBegs := make(map[[8]byte]time.Time)
	for {
		tm := time.Now()
		select {
		case entry := <-addCh:
			tBegs[entry.id] = entry.tBeg
		case id := <-delCh:
			delete(tBegs, id)
		case id := <-getReqCh:
			getRetCh <- tBegs[id]
		}
		timing.Since(tm, "CNVT Provider//convertRequest < EACH Req")
	}
} */

func tpConvert(converter *protocol.SimpleConverter, cReqMsg <-chan []byte, pReqMsg chan<- []byte) {
	var cpreq protocol.CustRequest
	for {
		tm := time.Now()
		msg := <-cReqMsg
		//log.Println("msg from cReqMsg", msg)
		cpreq.FromByteArr(msg)
		dpreq, err := converter.CustomToDubbo(cpreq)
		if err != nil {
			log.Fatal(err)
		}
		dbreq, err := dpreq.ToByteArr()
		if err != nil {
			log.Fatal(err)
		}

		err = dpreq.CheckFormat(dbreq)
		if err != nil {
			log.Fatal(err)
		}
		pReqMsg <- dbreq
		timing.Since(tm, "CNVT Provider//convertRequest < EACH Req")
	}
}
func providerWrite(pConn net.Conn, pReqMsg <-chan []byte) {
	for {
		tm := time.Now()
		dbReq := <-pReqMsg

		//log.Println("out", dbReq)
		n, err := pConn.Write(dbReq)

		if err != nil || n != len(dbReq) {
			log.Println(err)
			return
		}
		//log.Println("to provider")
		//log.Println(dbreq)
		timing.Since(tm, "WRIT Provider//providerWrite < EACH Req")
	}
}
func providerRead(pConn net.Conn, pRespMsg chan<- []byte) {
	for {
		tm := time.Now()
		dbh := make([]byte, 16)
		_, err := io.ReadFull(pConn, dbh)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println("Dubbo Head:", dbh)
		lens := binary.BigEndian.Uint32(dbh[12:16])
		dbrep := make([]byte, lens)
		_, err = io.ReadFull(pConn, dbrep)
		if err != nil {
			log.Println(err)
			return
		}
		dbrep = append(dbh, dbrep...)

		var id [8]byte
		copy(id[:], dbh[4:12])
		pRespMsg <- dbrep
		timing.Since(tm, "READ Provider//providerRead < EACH Req")
	}
}
func tcConvert(converter *protocol.SimpleConverter, pRespMsg <-chan []byte, cRespMsg chan<- []byte) {
	var dprep protocol.DubboPacks
	for {
		tm := time.Now()
		//log.Println("From provider:")
		//log.Println(dbrep)
		// dprep.FromByteArr(<-pRespMsg)
		msg := <-pRespMsg
		dprep.FromByteArr(msg)
		cprep, err := converter.DubboToCustom(uint64(0), dprep)
		//log.Println("msg", msg, cprep)
		if err != nil {
			// log.Fatal(err)
			// return
			log.Println(err)
			continue
		}
		cbrep, err := cprep.ToByteArr()
		if err != nil {
			// log.Fatal(err)
			// return
			log.Println(err)
			continue
		}

		cRespMsg <- cbrep
		timing.Since(tm, "CNVT Provider//tcConvert < EACH Req")
	}
}
func clientWrite(cConn net.Conn, cRespMsg <-chan []byte) {
	for {
		tm := time.Now()
		bl := make([]byte, 4)
		cbrep := <-cRespMsg
		binary.BigEndian.PutUint32(bl, uint32(len(cbrep)))

		_, err := cConn.Write(bl)
		if err != nil {
			log.Println(err)
			return
		}

		//log.Println("to customer", cbrep)
		_, err = cConn.Write(cbrep)
		if err != nil {
			log.Println(err)
			return
		}
		timing.Since(tm, "WRIT Provider//clientWrite < EACH Req")
	}
}

//Stop can be used for closing provider manually.
func (p *Provider) Stop() {
	p.chanStop <- nil
}

//keepAlive receive the etcdv3.response, and update lease.
func (p *Provider) keepAlive() <-chan *etcdv3.LeaseKeepAliveResponse {
	log.Println("Ready to keepAlive!")

	info := &p.info

	key := p.name
	value, _ := json.Marshal(info)

	resp, err := p.client.Grant(context.Background(), MinTTL)
	if err != nil {
		log.Fatal(err)
	}

	_, err = p.client.Put(context.Background(), key, string(value), etcdv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
	}
	p.leaseId = resp.ID

	log.Println("Put OK!", key, string(value))

	ret, err := p.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Keep Alive OK!")

	return ret
}

//revoke is a wrapper of etcd.Revoke.
func (p *Provider) revoke() error {

	_, err := p.client.Revoke(context.Background(), p.leaseId)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
