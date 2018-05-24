package consumer

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"log"
	"net/http"
	"io"
	"protocol"
	"math"
	"encoding/binary"
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
	}

	go c.start()
	go c.listen()

	return c
}

func (c *Connection) write() {
	lb := make([]byte, 4)
	lens := uint32(0)
	for {
		cpreq := <-c.provider.chanIn
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatal(err)
			return
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(lb, lens)
		c.conn.Write(append(lb, cbreq...))
	}
}

func (c *Connection) read() {
	lb := make([]byte, 4)
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
		c.consumer.answer[cprep.Identifier] <- cprep.Reply
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

	var hp protocol.HttpPacks

	hp.FromRequests(r)

	cpreq, err := c.cnvt.HTTPToCustom(hp)
	if err != nil {
		log.Fatal(err)
		return
	}

	id := cpreq.Identifier
	c.providers[minDelayId].chanIn <- cpreq

	c.answer[id] = make(chan []byte)

	io.WriteString(w, string(<-c.answer[id]))
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
