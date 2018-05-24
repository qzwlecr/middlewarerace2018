package provider

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"protocol"
	"time"

	etcdv3 "github.com/coreos/etcd/clientv3"
)

// change it to your listen address and provider address
const (
	lnAddr       = ":30000"
	providerAddr = ":20880"
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

func handleReq(ln net.Listener, tcpCh <-chan int, converter *protocol.SimpleConverter) {
	defer ln.Close()

	go func(converter *protocol.SimpleConverter) {

		for {
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
			pReqMsg := make(chan []byte, 10)
			cRespMsg := make(chan []byte, 10)
			pRespMsg := make(chan []byte, 10)
			elapsedCh := make(chan int64, 10)
			go func(cConn net.Conn, cReqMsg chan<- []byte) {
				for {
					bl := make([]byte, 4)
					io.ReadFull(cConn, bl)
					lens := binary.BigEndian.Uint32(bl)

					cbreq := make([]byte, lens)
					io.ReadFull(cConn, cbreq)

					cReqMsg <- cbreq
				}
			}(cConn, cReqMsg)

			// maintaing timing map
			type TMapEntry struct {
				id   [8]byte
				tBeg time.Time
			}

			addCh := make(chan TMapEntry, 5)
			delCh := make(chan [8]byte, 5)
			getReqCh := make(chan [8]byte, 5)
			getRetCh := make(chan time.Time, 5)
			go func(addCh <-chan TMapEntry, delCh, getReqCh <-chan [8]byte, getRetCh chan<- time.Time) {
				tBegs := make(map[[8]byte]time.Time)
				for {
					select {
					case entry := <-addCh:
						tBegs[entry.id] = entry.tBeg
					case id := <-delCh:
						delete(tBegs, id)
					case id := <-getReqCh:
						getRetCh <- tBegs[id]
					}
				}
			}(addCh, delCh, getReqCh, getRetCh)

			// to provider converter
			go func(converter *protocol.SimpleConverter, cReqMsg <-chan []byte, pReqMsg chan<- []byte, addCh chan<- TMapEntry) {
				var cpreq protocol.CustRequest
				for {
					cpreq.FromByteArr(<-cReqMsg)

					timingBeg := time.Now()

					dpreq, err := converter.CustomToDubbo(cpreq)
					if err != nil {
						log.Fatal(err)
					}
					dbreq, err := dpreq.ToByteArr()
					if err != nil {
						log.Fatal(err)
					}

					var entry TMapEntry
					entry.tBeg = timingBeg
					copy(entry.id[:], dbreq[4:12])
					addCh <- entry

					pReqMsg <- dbreq
				}
			}(converter, cReqMsg, pReqMsg, addCh)

			// to server write
			go func(pConn net.Conn, pReqMsg <-chan []byte) {
				for {
					dbReq := <-pReqMsg

					n, err := pConn.Write(dbReq)

					if err != nil || n != len(dbReq) {
						log.Println(err)
						return
					}
					//log.Println("to provider")
					//log.Println(dbreq)
				}
			}(pConn, pReqMsg)

			// from server read
			go func(pConn net.Conn, pRespMsg chan<- []byte, getReqCh chan<- [8]byte, getRetCh <-chan time.Time, elapsedCh chan int64) {
				for {
					dbh := make([]byte, 16)
					_, err := io.ReadFull(pConn, dbh)
					if err != nil {
						log.Println(err)
						return
					}
					//log.Println("Dubbo Head:", dbh)
					lens := binary.BigEndian.Uint32(dbh[12:16])
					dbrep := make([]byte, lens)
					io.ReadFull(pConn, dbrep)
					dbrep = append(dbh, dbrep...)

					var id [8]byte
					copy(id[:], dbh[4:12])
					getReqCh <- id
					timingBeg := <-getRetCh

					timingEnd := time.Now()
					elapsed := timingEnd.Sub(timingBeg).Nanoseconds() / 1000

					elapsedCh <- elapsed
					pRespMsg <- dbrep
				}
			}(pConn, pRespMsg, getReqCh, getRetCh, elapsedCh)

			// from provider converter
			go func(converter *protocol.SimpleConverter, pRespMsg <-chan []byte, cRespMsg chan<- []byte, elapsedCh <-chan int64) {
				var dprep protocol.DubboPacks
				for {
					//log.Println("From provider:")
					//log.Println(dbrep)
					elapsed := <-elapsedCh
					dprep.FromByteArr(<-pRespMsg)
					cprep, err := converter.DubboToCustom(uint64(elapsed), dprep)
					if err != nil {
						log.Fatal(err)
						return
					}
					cbrep, err := cprep.ToByteArr()
					if err != nil {
						log.Fatal(err)
						return
					}

					cRespMsg <- cbrep
				}
			}(converter, pRespMsg, cRespMsg, elapsedCh)

			// to client write
			go func(cConn net.Conn, cRespMsg <-chan []byte) {
				for {
					bl := make([]byte, 4)
					cbrep := <-cRespMsg
					binary.BigEndian.PutUint32(bl, uint32(len(cbrep)))

					_, err = cConn.Write(bl)
					if err != nil {
						log.Println(err)
						return
					}

					//log.Println("To customer")
					//log.Println(cbrep)
					_, err = cConn.Write(cbrep)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}(cConn, cReqMsg)
		}
	}(converter)

	<-tcpCh
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
