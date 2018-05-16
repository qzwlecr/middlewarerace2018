package consumer

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"log"
	"context"
	"encoding/json"
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
		etcdAddr: endpoints,
		providers: make(map[string]*Provider),
		client:    cli,
	}

	go c.Start()

	return c
}

//Start shouldn't be called manually.
func (c *Consumer) Start() {
	c.watchProvider()
}

//Stop must be used for closing connection.
func (c *Consumer) Stop() {
	c.client.Close()
}

//addProvider add (key,info) to the consumer's map.
func (c *Consumer) addProvider(key string, info *ProviderInfo) {
	p := &Provider{
		name: key,
		info: *info,
	}
	c.providers[p.name] = p
}

//getProviderInfo return one etcdv3.event's info(Marshaled by Json).
func getProviderInfo(ev *etcdv3.Event) *ProviderInfo {
	info := &ProviderInfo{}
	err := json.Unmarshal([]byte(ev.Kv.Value), info)
	if err != nil {
		log.Fatal(err)
	}
	return info
}

//watchProvider can auto update the consumer's provider-map.
func (c *Consumer) watchProvider() {
	chanWatch := c.client.Watch(context.Background(), c.path, etcdv3.WithPrefix())
	for wresp := range chanWatch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case etcdv3.EventTypePut:
				info := getProviderInfo(ev)
				log.Println(string(ev.Kv.Key) + " " + info.IP + " is Connecting!")
				c.addProvider(string(ev.Kv.Key), info)
			case etcdv3.EventTypeDelete:
				log.Println(string(ev.Kv.Key) + " Has Been removed!")
				delete(c.providers, string(ev.Kv.Key))
			}
		}
	}

}
