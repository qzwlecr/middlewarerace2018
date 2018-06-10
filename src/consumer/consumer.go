package consumer

import (
	"io"
	"log"
	"net/http"
	"protocol"
	"time"
	"utility/timing"

	"math"
	"sync"

	etcdv3 "github.com/coreos/etcd/clientv3"
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
	//defer timing.Since(time.Now(), "[INFO]Client handling:")
	if len(c.providers) == 0 {
		return
	}

	tm := time.Now()

	chosenId := c.chooseProvider()

	var hp protocol.HttpPacks

	hp.FromRequests(r)

	cpreq, err := c.cnvt.HTTPToCustom(hp)
	if err != nil {
		log.Fatal(err)
		return
	}

	id := cpreq.Identifier

	chanByte := make(chan []byte)
	ti := time.Now()

	c.answer.LoadOrStore(id, chanByte)
	if logger {
		log.Println("[INFO]Using provider:", chosenId, "  ", c.providers[chosenId].delay)
	}
	go func() {
		c.providers[chosenId].chanIn <- cpreq
	}()

	timing.Since(tm, "Procedure time: ")
	//defer timing.Since(time.Now(), "[INFO]Request has been sent.")

	ans := string(<-chanByte)
	go func(t time.Time) {
		c.providers[chosenId].chanTime <- t
	}(ti)

	io.WriteString(w, ans)
}

func (c *Consumer) listen() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}

func (c *Consumer) chooseProvider() string {
	// test trigger rebuilt.
	minDelay := int64(math.MaxInt64)
	minDelayId := ""
	for id, p := range c.providers {
		//log.Println(p.info.IP, "Active: ", p.active, ", Delay: ", p.delay)
		if p.delay < minDelay {
			minDelayId = id
			minDelay = p.delay
		}
	}
	if minDelayId == "" {
		log.Panic("c.providers boom!")
	}
	return minDelayId
}
