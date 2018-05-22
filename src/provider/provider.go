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

			go func(cConn net.Conn, converter *protocol.SimpleConverter) {
				defer cConn.Close()

				pConn, err := net.Dial("tcp", providerAddr)
				if err != nil {
					log.Fatal(err)
				}

				defer pConn.Close()

				bl := make([]byte, 4)
				io.ReadFull(cConn, bl)
				lens := binary.BigEndian.Uint32(bl)

				cbreq := make([]byte, lens)
				io.ReadFull(cConn, cbreq)
				//log.Println("from customer")
				//log.Println(cbreq)

				var cpreq protocol.CustRequest

				timingBeg := time.Now()

				cpreq.FromByteArr(cbreq)
				dpreq := converter.CustomToDubbo(cpreq)
				dbreq := dpreq.ToByteArr()

				n, err := pConn.Write(dbreq)

				if err != nil || n != len(dbreq) {
					log.Println(err)
					return
				}

				//log.Println("to provider")
				//log.Println(dbreq)

				dbh := make([]byte, 12)
				io.ReadFull(pConn, dbh)
				//log.Println("Dubbo Head:", dbh)
				lens = binary.LittleEndian.Uint32(dbh[8:12])
				dbrep := make([]byte, lens)
				io.ReadFull(pConn, dbrep)
				dbrep = append(dbh, dbrep...)

				timingEnd := time.Now()
				elapsed := timingEnd.Sub(timingBeg).Nanoseconds() / 1000

				var dprep protocol.DubboPacks
				//log.Println("from provider")
				//log.Println(dbrep)
				dprep.FromByteArr(dbrep)
				cprep := converter.DubboToCustom(uint64(elapsed), dprep)
				cbrep := cprep.ToByteArr()

				binary.BigEndian.PutUint32(bl, uint32(len(cbrep)))

				_, err = cConn.Write(bl)
				if err != nil {
					log.Println(err)
					return
				}

				//log.Println("to customer")
				//log.Println(cbrep)
				_, err = cConn.Write(cbrep)
				if err != nil {
					log.Println(err)
					return
				}
			}(cConn, converter)
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
