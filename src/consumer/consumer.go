package consumer

import (
	"io"
	"log"
	"net/http"
	"protocol"
	"time"
	"utility/timing"

	etcdv3 "github.com/coreos/etcd/clientv3"
	"math"
	"sync"
)

const (
	listenPort  = "20000"
	requestPort = "30000"
)

type Consumer struct {
	path      string
	etcdAddr  []string
	cnvt      protocol.SimpleConverter
	answer    sync.Map
	providers map[string]*Provider
	client    *etcdv3.Client
}

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

	chosenId := c.chooseProvider()

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
	if logger {
		log.Println("[INFO]Using provider:", chosenId, "  ", c.providers[chosenId].info.IP)
	}
	c.providers[chosenId].chanIn <- cpreq

	defer timing.Since(time.Now(), "[INFO]Request has been sent.")

	io.WriteString(w, string(<-ch))
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}

func (c *Consumer) chooseProvider() string {

	minDelay := uint64(math.MaxUint32)
	minDelayId := ""
	minActive := uint32(math.MaxUint32)
	for id, p := range c.providers {
		log.Println(p.info.IP, "Active: ", p.active, ", Delay: ", p.delay)
		if p.delay < minDelay && p.active/p.weight < minActive {
			minDelayId = id
			minDelay = p.delay
			minActive = p.active / p.weight
		}
	}
	if minDelayId == "" {
		log.Panic("c.providers boom!")
	}
	return minDelayId
}
