package consumer

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"net/http"
	"protocol"
	"time"
	"utility/timing"

	etcdv3 "github.com/coreos/etcd/clientv3"
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

func (c Connection) write() {
	lb := make([]byte, 4)
	lens := uint32(0)
	for {
		select {
		case cpreq := <-c.provider.chanIn:
			ti := time.Now()
			cbreq, err := cpreq.ToByteArr()
			if err != nil {
				log.Fatal(err)
				return
			}

			lens = uint32(len(cbreq))
			binary.BigEndian.PutUint32(lb, lens)
			fullp := append(lb, cbreq...)
			//log.Println("Write Packages:", fullp)
			c.conn.Write(fullp)
			timing.Since(ti, "[INFO]Writing: ")
		}
	}
}

func (c Connection) read() {
	lb := make([]byte, 4)
	if c.conn == nil {
		log.Panic("Conn boom in reader!")
	}
	//log.Println(c.conn.LocalAddr())
	for {
		n, err := io.ReadFull(c.conn, lb)
		if n != 4 || err != nil {
			log.Fatal(err)
			return
		}
		ti := time.Now()
		lens := binary.BigEndian.Uint32(lb)
		cbrep := make([]byte, lens)
		n, err = io.ReadFull(c.conn, cbrep)

		if n != int(lens) || err != nil {
			log.Fatal(err)
			return
		}
		var cprep protocol.CustResponse
		cprep.FromByteArr(cbrep)
		//log.Println("Read Packages:", cbrep)
		//log.Println("Writing answers to map:")
		ch, _ := c.consumer.answer.Load(cprep.Identifier)
		go func() { ch.(chan []byte) <- cprep.Reply }()
		//log.Println("Writing answers to map Done.")
		c.provider.delay = (c.provider.delay + cprep.Delay) / 2
		timing.Since(ti, "[INFO]Reading: ")
	}
}

//Start shouldn't be called manually.
func (c *Consumer) start() {
	defer c.client.Close()
	c.watchProvider()
}

func (c *Consumer) clientHandler(w http.ResponseWriter, r *http.Request) {
	defer timing.Since(time.Now(), "[INFO]Client handling:")
	if len(c.providers) == 0 {
		return
	}

	minDelay := uint64(math.MaxUint64)
	minDelayId := ""
	for id, p := range c.providers {
		//log.Println("[INFO]Providers info:", p.info.IP, "  ", p.delay)
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

	ch := make(chan []byte)
	c.answer.LoadOrStore(id, ch)

	log.Println("[INFO]Using provider:", minDelayId, "  ", c.providers[minDelayId].info.IP)
	c.providers[minDelayId].chanIn <- cpreq

	defer timing.Since(time.Now(), "[INFO]Request has been sent.")

	//log.Println("Waiting for reading.")

	io.WriteString(w, string(<-ch))
	//log.Println("all things have been done.")
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
