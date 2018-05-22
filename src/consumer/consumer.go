package consumer

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"log"
	"context"
	"encoding/json"
	"net/http"
	"io"
	"protocol"
	"math"
	"net"
	"encoding/binary"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
		chanProv:  make(chan bool),
		providers: make(map[string]*Provider),
		client:    cli,
	}

	go c.Start()
	go c.communicate()

	return c
}

//Start shouldn't be called manually.
func (c *Consumer) Start() {
	defer c.client.Close()
	c.watchProvider()
}

//addProvider add (key,info) to the consumer's map.
func (c *Consumer) addProvider(key string, info *ProviderInfo) {
	p := &Provider{
		name:  key,
		info:  *info,
		delay: 0,
	}
	c.providers[p.name] = p
	//log.Println("Some provider comes in!")
}

//getProviderInfo return one etcdv3.event's info(Marshaled by Json).
func getProviderInfo(kv mvccpb.KeyValue) *ProviderInfo {
	info := &ProviderInfo{}
	err := json.Unmarshal([]byte(kv.Value), info)
	if err != nil {
		log.Fatal(err)
	}
	return info
}

//watchProvider can auto update the consumer's provider-map.
func (c *Consumer) watchProvider() {
	//log.Println("Watching provider!")
	rangeResp, err := c.client.Get(context.TODO(), c.path, etcdv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, kv := range rangeResp.Kvs {
		info := getProviderInfo(*kv)
		//log.Println(string(kv.Key) + " " + info.IP + " is Connecting!")
		c.addProvider(string(kv.Key), info)

	}
	//defer log.Println("Stop watching!")
	chanWatch := c.client.Watch(context.Background(), c.path, etcdv3.WithPrefix())
	for wresp := range chanWatch {
		for _, ev := range wresp.Events {
			//log.Println("Some event happens!")
			switch ev.Type {
			case etcdv3.EventTypePut:
				info := getProviderInfo(*ev.Kv)
				//log.Println(string(ev.Kv.Key) + " " + info.IP + " is Connecting!")
				c.addProvider(string(ev.Kv.Key), info)
			case etcdv3.EventTypeDelete:
				//log.Println(string(ev.Kv.Key) + " Has Been removed!")
				delete(c.providers, string(ev.Kv.Key))
			}
		}
	}

}

func (c *Consumer) clientHandler(w http.ResponseWriter, r *http.Request) {
	for len(c.providers) == 0 {

	}
	cnvt := new(protocol.SimpleConverter)
	hp := new(protocol.HttpPacks)

	hp.FromRequests(r)

	//log.Println("HTTP Pack Payload:", hp.Payload)

	cpreq := cnvt.HTTPToCustom(*hp)
	cbreq := cpreq.ToByteArr()

	minDelay := uint64(math.MaxUint64)
	minDelayId := ""
	for id, p := range c.providers {
		if minDelay > p.delay {
			minDelayId = id
			minDelay = p.delay
		}
	}

	d, err := net.Dial("tcp", net.JoinHostPort(c.providers[minDelayId].info.IP, requestPort))
	if err != nil {
		log.Fatal(err)
	}
	lens := uint32(len(cbreq))
	lb := make([]byte, 4)
	binary.BigEndian.PutUint32(lb, lens)
	d.Write(lb)
	d.Write(cbreq)
	io.ReadFull(d, lb)
	lens = binary.BigEndian.Uint32(lb)
	//log.Println("Reply length:", lens)
	cbrep := make([]byte, lens)
	io.ReadFull(d, cbrep)
	//log.Println("Reply content:", cbrep)
	cprep := new(protocol.CustResponse)
	cprep.FromByteArr(cbrep)
	c.providers[minDelayId].delay = (c.providers[minDelayId].delay + cprep.Delay) / 2

	*hp = cnvt.CustomToHTTP(*cprep)

	hb := hp.ToByteArr()
	//log.Println("HTTP reply:", hb)
	io.WriteString(w, string(hb))
}

func (c *Consumer) communicate() {
	http.HandleFunc("/", c.clientHandler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
