package consumer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"protocol"
	"sync"
	"time"

	etcdv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

var connId int

type Consumer struct {
	watchPath     string
	etcdAddr      []string
	etcdClient    *etcdv3.Client
	providers     map[string]*provider
	converter     protocol.SimpleConverter
	connections   []*connection
	connectionsMu sync.Mutex
	answerMu      sync.RWMutex
	answer        map[uint64]chan answer
	chanOut       chan protocol.CustRequest
	chanIn        chan answer
}

type answer struct {
	connId int
	id     uint64
	reply  []byte
}

//NewConsumer receive etcd server address, and the services watchPath on etcd.
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
		watchPath:   watchPath,
		etcdAddr:    endpoints,
		etcdClient:  cli,
		providers:   make(map[string]*provider),
		converter:   protocol.SimpleConverter{},
		connections: make([]*connection, 0),
		chanOut:     make(chan protocol.CustRequest, queueSize),
		chanIn:      make(chan answer, queueSize),
		answer:      make(map[uint64]chan answer),
	}

	go c.watchProvider()
	go c.listenHTTP()
	go c.updateAnswer()
	go c.OverloadCheck()
	return c
}

func (c *Consumer) OverloadCheck() {
	for {
		<-time.After(25 * time.Millisecond)
		if len(c.chanOut) > overLoadSize {
			if logger {
				log.Println("Overload!")
			}
			go c.overload()
		}
		if len(c.connections) > connMaxSize {
			return
		}
	}
}

func (c *Consumer) clientHandler(w http.ResponseWriter, r *http.Request) {
	for len(c.providers) == 0 {
	}

	var hp protocol.HttpPacks
	hp.FromRequests(r)

	cpreq, err := c.converter.HTTPToCustom(hp)
	if err != nil {
		log.Fatalln(err)
		return
	}
	ch := make(chan answer)
	id := cpreq.Identifier

	c.answerMu.Lock()
	c.answer[id] = ch
	c.answerMu.Unlock()

	c.chanOut <- cpreq

	t := time.Now()

	ret := <-ch

	delay := time.Since(t)
	connection := c.connections[ret.connId]
	provider := connection.provider
	connection.ignoreNum++
	if connection.ignoreNum > ignoreSize {
		go func(duration time.Duration) {
			log.Println(connection.provider.info, "has Delay:", delay)
			provider.chanDelay <- delay
		}(delay)
	}

	io.WriteString(w, string(ret.reply))
}

func (c *Consumer) listenHTTP() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}

//addProvider add (key,info) to the consumer's map.
func (c *Consumer) addProvider(key string, info providerInfo) {
	p := &provider{
		name:            key,
		info:            info,
		baseDelay:       0,
		baseDelaySample: 0,
		weight:          info.Weight,
		consumer:        c,
		connectionSize:  0,
		fullLevel:       0,
		isFull:          false,
		chanDelay:       make(chan time.Duration, queueSize),
	}
	c.providers[p.name] = p

	for i := 0; i < connMinSize; i++ {
		c.addConnection(p)
	}
	//if p.name == "/provider/small" {
	//	for i := 0; i < 96; i++ {
	//		c.addConnection(p)
	//	}
	//} else {
	//	if p.name == "/provider/medium" {
	//		for i := 0; i < 192; i++ {
	//			c.addConnection(p)
	//		}
	//	} else {
	//		for i := 0; i < 224; i++ {
	//			c.addConnection(p)
	//		}
	//
	//	}
	//}
	//p.tryConnect()

	go p.maintain()
}

func (c *Consumer) addConnection(p *provider) {
	connection := &connection{
		consumer: c,
		provider: p,
	}
	c.connectionsMu.Lock()
	connection.connId = len(c.connections)
	c.connections = append(c.connections, connection)
	c.connectionsMu.Unlock()
	conn, _ := net.Dial("tcp", net.JoinHostPort(p.info.IP, requestPort))
	go connection.dealWithConnection(conn)
}

//getProviderInfo return one etcdv3.event's info(Marshaled by Json).
func (c *Consumer) getProviderInfo(kv mvccpb.KeyValue) providerInfo {
	info := providerInfo{}
	err := json.Unmarshal([]byte(kv.Value), &info)
	if err != nil {
		log.Fatal(err)
	}
	return info
}

//watchProvider can auto update the consumer's provider-map.
func (c *Consumer) watchProvider() {
	rangeResp, err := c.etcdClient.Get(context.TODO(), c.watchPath, etcdv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, kv := range rangeResp.Kvs {
		info := c.getProviderInfo(*kv)
		c.addProvider(string(kv.Key), info)
	}
	chanWatch := c.etcdClient.Watch(context.Background(), c.watchPath, etcdv3.WithPrefix())
	for wresp := range chanWatch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case etcdv3.EventTypePut:
				info := c.getProviderInfo(*ev.Kv)
				c.addProvider(string(ev.Kv.Key), info)
			case etcdv3.EventTypeDelete:
				delete(c.providers, string(ev.Kv.Key))
			}
		}
	}
}

func (c *Consumer) updateAnswer() {
	for ans := range c.chanIn {
		c.answerMu.RLock()
		c.answer[ans.id] <- ans
		c.answerMu.RUnlock()
	}
}

func (c *Consumer) overload() {
	for _, p := range c.providers {
		if p.isFull == false {
			c.addConnection(p)
			if logger {
				log.Println(p.info, "Now have new connection. Total:", len(c.connections))
			}
			return
		}
	}

}
