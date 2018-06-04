package consumer

import (
	"encoding/binary"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"io"
	"log"
	"math"
	"net/http"
	"protocol"
	"sync"
)

const (
	listenPort  = "20000"
	requestPort = "30000"
)

//NewConsumer receive etcd server address, and the services path on etcd.
//And return the consumer which has been started.
func NewConsumer(endpoints []string, watchPath string) *Consumer {
	cfg := etcdv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	cli, err := etcdv3.New(cfg)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	c := &Consumer{
		path:      watchPath,
		etcdAddr:  endpoints,
		providers: make(map[string]*Provider),
		client:    cli,
		answer:    make(map[uint64]chan []byte),
		answerMu:  sync.RWMutex{},
	}

	go c.start()
	go c.listen()

	return c
}

func (c Connection) write() {
	lb := make([]byte, 4)
	lens := uint32(0)
	for {
		select {
		case cpreq := <-c.provider.chanIn:
			cbreq, err := cpreq.ToByteArr()
			if err != nil {
				log.Fatal(err)
				return
			}

			lens = uint32(len(cbreq))
			binary.BigEndian.PutUint32(lb, lens)
			fullp := append(lb, cbreq...)
			log.Println("Write Packages:", fullp)
			c.conn.Write(fullp)
		}
	}
}

func (c Connection) read() {
	lb := make([]byte, 4)
	if c.conn == nil {
		log.Panic("Conn boom in reader!")
	}
	log.Println(c.conn.LocalAddr())
	for {
		n, err := io.ReadFull(c.conn, lb)
		if n != 4 || err != nil {
			log.Fatal(err)
			return
		}
		lens := binary.BigEndian.Uint32(lb)
		cbrep := make([]byte, lens)
		n, err = io.ReadFull(c.conn, cbrep)

		if n != int(lens) || err != nil {
			log.Fatal(err)
			return
		}
		var cprep protocol.CustResponse
		cprep.FromByteArr(cbrep)
		log.Println("Read Packages:", cbrep)
		c.consumer.answerMu.Lock()
		c.consumer.answer[cprep.Identifier] <- cprep.Reply
		c.consumer.answerMu.Unlock()
		c.provider.delay = (c.provider.delay + cprep.Delay) / 2
	}
}

//Start shouldn't be called manually.
func (c *Consumer) start() {
	defer c.client.Close()
	c.watchProvider()
}

func (c *Consumer) clientHandler(w http.ResponseWriter, r *http.Request) {
	for len(c.providers) == 0 {

	}

	minDelay := uint64(math.MaxUint64)
	minDelayId := ""
	for id, p := range c.providers {
		if minDelay > p.delay {
			minDelayId = id
			minDelay = p.delay
		}
	}
	if minDelayId == "" {
		log.Panic("c.providers boom!")
	}

	var hp protocol.HttpPacks

	hp.FromRequests(r)

	cpreq, err := c.cnvt.HTTPToCustom(hp)
	if err != nil {
		log.Fatal(err)
		return
	}

	id := cpreq.Identifier
	c.providers[minDelayId].chanIn <- cpreq

	c.answerMu.Lock()
	c.answer[id] = make(chan []byte)
	c.answerMu.Unlock()

	io.WriteString(w, string(<-c.answer[id]))
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
