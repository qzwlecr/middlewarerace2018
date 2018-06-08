package consumer

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"protocol"
	"utility/timing"
	"time"

	etcdv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

type Provider struct {
	name     string
	etcdAddr []string
	delay    uint64
	info     ProviderInfo
	leaseId  etcdv3.LeaseID
	client   *etcdv3.Client
	chanIn   chan Mission
	conns    []Connection
	active   uint32
	weight   uint32
}

type Mission struct {
	cr         protocol.CustRequest
	chanAnswer chan []byte
}

//addProvider add (key,info) to the consumer's map.
func (c *Consumer) addProvider(key string, info ProviderInfo) {
	defer timing.Since(time.Now(), "[INFO]Add Provider:")
	p := &Provider{
		name:   key,
		info:   info,
		delay:  0,
		weight: info.Weight,
		active: 0,
		chanIn: make(chan Mission, queueSize),
		conns:  make([]Connection, connsSize),
	}
	for _, ec := range p.conns {
		ec.consumer = c
		ec.provider = p
		ec.isActive = false
		ec.answer = new(sync.Map)
		conn, err := net.Dial("tcp", net.JoinHostPort(info.IP, requestPort))

		if err != nil {
			log.Fatal(err)
		}
		if conn == nil {
			log.Panic("Conn boom in provider!")
		}
		go ec.read(conn)
		go ec.write(conn)
	}
	c.providers[p.name] = p
}

//getProviderInfo return one etcdv3.event's info(Marshaled by Json).
func getProviderInfo(kv mvccpb.KeyValue) ProviderInfo {
	info := ProviderInfo{}
	err := json.Unmarshal([]byte(kv.Value), &info)
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
