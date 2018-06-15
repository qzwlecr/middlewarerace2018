package provider

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"protocol"
	"time"
	"utility/timing"

	_ "net/http/pprof"

	etcdv3 "github.com/coreos/etcd/clientv3"
)

// change it to your listen address and provider address
const (
	lnAddr       = ":30000"
	providerAddr = ":20880"
	lnPpofAddr   = ":40000"
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
	}

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
	http.ListenAndServe(lnPpofAddr, nil)

	tcpCh := make(chan int)

	var converter protocol.SimpleConverter
	// handler for listening over tcp
	go handleReq(ln, tcpCh, &converter)

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

func handleReq(ln net.Listener, tcpCh <-chan int, converter *protocol.SimpleConverter) {
	defer ln.Close()

	go func(converter *protocol.SimpleConverter) {

		for {
			tm := time.Now()
			cConn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}

			// cConn.Close()

			pConn, err := net.Dial("tcp", providerAddr)
			if err != nil {
				log.Fatal(err)
			}
			// pConn.Close()

			// from client read
			cReqMsg := make(chan []byte, 10)
			// pReqMsg := make(chan []byte, 10)
			// cRespMsg := make(chan []byte, 10)
			pRespMsg := make(chan []byte, 10)
			elapsedCh := make(chan int64, 10)
			go clientRead(cConn, cReqMsg)

			addCh := make(chan tMapEntry, 5)
			delCh := make(chan [8]byte, 5)
			getReqCh := make(chan [8]byte, 1)
			getRetCh := make(chan time.Time, 1)
			go convertRequest(addCh, delCh, getReqCh, getRetCh)

			// to provider converter
			// go tpConvert(converter, cReqMsg, pReqMsg, addCh)

			// to server write
			go providerWrite(converter, cReqMsg, addCh, pConn)

			// from server read
			go providerRead(pConn, pRespMsg, getReqCh, delCh, getRetCh, elapsedCh)

			// from provider converter
			// go tcConvert(converter, pRespMsg, cRespMsg, elapsedCh)

			// to client write
			go clientWrite(converter, pRespMsg, elapsedCh, cConn)
			timing.Since(tm, "HAND Provider//HandleReq < EACH")
		}
	}(converter)

	<-tcpCh
}

func clientRead(cConn net.Conn, cReqMsg chan<- []byte) {
	defer func(cConn net.Conn) {
		cConn.Close()
		close(cReqMsg)
	}(cConn)

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

		//log.Println("msg to cReqMsg", cbreq)
		cReqMsg <- cbreq
		timing.Since(tm, "READ Provider//clientRead < EACH Req")
	}
}
func convertRequest(addCh <-chan tMapEntry, delCh, getReqCh <-chan [8]byte, getRetCh chan<- time.Time) {
	tBegs := make(map[[8]byte]time.Time)
	for {
		tm := time.Now()
		select {
		case entry, moreAdd := <-addCh:
			tBegs[entry.id] = entry.tBeg
			if !moreAdd {
				addCh = nil
			}
		case id, moreDel := <-delCh:
			delete(tBegs, id)
			if !moreDel {
				delCh = nil
			}
		case id, moreGet := <-getReqCh:
			getRetCh <- tBegs[id]
			if !moreGet {
				getReqCh = nil
			}
		}
		timing.Since(tm, "CNVT Provider//convertRequest < EACH Req")
		if addCh == nil && delCh == nil && getReqCh == nil {
			break
		}
	}
}

// func tpConvert(converter *protocol.SimpleConverter, cReqMsg <-chan []byte, pReqMsg chan<- []byte, addCh chan<- tMapEntry) {
// }

func providerWrite(converter *protocol.SimpleConverter, cReqMsg <-chan []byte, addCh chan<- tMapEntry, pConn net.Conn) {
	defer func(pConn net.Conn) {
		pConn.Close()
		close(addCh)
	}(pConn)

	var cpreq protocol.CustRequest
	for {
		tm := time.Now()
		msg, more := <-cReqMsg
		if !more {
			return
		}

		//log.Println("msg from cReqMsg", msg)
		cpreq.FromByteArr(msg)

		timingBeg := time.Now()

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

		var entry tMapEntry
		entry.tBeg = timingBeg
		copy(entry.id[:], dbreq[4:12])
		addCh <- entry

		// pReqMsg <- dbreq

		//log.Println("out", dbReq)
		n, err := pConn.Write(dbreq)
		log.Println("current requests pending: ", len(cReqMsg))

		if err != nil || n != len(dbreq) {
			log.Println(err)
			return
		}
		//log.Println("to provider")
		//log.Println(dbreq)
		timing.Since(tm, "WRIT Provider//providerWrite < EACH Req")
	}
}
func providerRead(pConn net.Conn, pRespMsg chan<- []byte, getReqCh, delCh chan<- [8]byte, getRetCh <-chan time.Time, elapsedCh chan int64) {
	defer func(pRespMsg chan<- []byte, getReqCh, delCh chan<- [8]byte) {
		close(pRespMsg)
		close(getReqCh)
		close(delCh)
	}(pRespMsg, getReqCh, delCh)
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
		getReqCh <- id
		timingBeg := <-getRetCh
		delCh <- id

		timingEnd := time.Now()
		elapsed := timingEnd.Sub(timingBeg).Nanoseconds() / 1000

		log.Println("provider time elapsed: ", elapsed)

		elapsedCh <- elapsed
		pRespMsg <- dbrep
		timing.Since(tm, "READ Provider//providerRead < EACH Req")
	}
}

// func tcConvert(converter *protocol.SimpleConverter, pRespMsg <-chan []byte, cRespMsg chan<- []byte, elapsedCh <-chan int64) {
// }
func clientWrite(converter *protocol.SimpleConverter, pRespMsg <-chan []byte, elapsedCh <-chan int64, cConn net.Conn) {
	var dprep protocol.DubboPacks
	for {
		tm := time.Now()
		//log.Println("From provider:")
		//log.Println(dbrep)
		msg, more := <-pRespMsg
		if !more {
			return
		}
		elapsed := <-elapsedCh
		// dprep.FromByteArr(<-pRespMsg)
		dprep.FromByteArr(msg)
		cprep, err := converter.DubboToCustom(uint64(elapsed), dprep)
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

		bl := make([]byte, 4)
		binary.BigEndian.PutUint32(bl, uint32(len(cbrep)))

		_, err = cConn.Write(bl)
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
