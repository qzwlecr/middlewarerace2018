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
	//TODO: place read and write here
	//go c.writeToProvider()
	//go c.readFromProvider()

	return c
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

	var cnvt protocol.SimpleConverter
	var hp protocol.HttpPacks
	var cprep protocol.CustResponse
	lb := make([]byte, 4)

	hp.FromRequests(r)

	cpreq, err := cnvt.HTTPToCustom(hp)
	if err != nil {
		log.Fatal(err)
		return
	}

	cbreq, err := cpreq.ToByteArr()
	if err != nil {
		log.Fatal(err)
		return
	}

	d := c.providers[minDelayId].connection

	lens := uint32(len(cbreq))
	binary.BigEndian.PutUint32(lb, lens)
	d.Write(lb)
	d.Write(cbreq)

	io.ReadFull(d, lb)
	lens = binary.BigEndian.Uint32(lb)
	cbrep := make([]byte, lens)
	io.ReadFull(d, cbrep)

	cprep.FromByteArr(cbrep)

	c.providers[minDelayId].delay = (c.providers[minDelayId].delay + cprep.Delay) / 2

	hp, err = cnvt.CustomToHTTP(cprep)
	if err != nil {
		log.Fatal(err)
		return
	}

	hb, err := hp.ToByteArr()
	if err != nil {
		log.Fatal(err)
		return
	}

	io.WriteString(w, string(hb))
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
