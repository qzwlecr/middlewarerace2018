package discovery

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"context"
	"encoding/json"
)

func NewConsumer(endpoints []string, watchPath string) *Consumer {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	cli, err := clientv3.New(cfg)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	c := &Consumer{
		path:      watchPath,
		providers: make(map[string]*Provider),
		client:    cli,
	}

	defer cli.Close()
	go c.WatchProvider()
	return c
}

func (c *Consumer) AddProvider(key string, info *ProviderInfo) {
	p := &Provider{
		name: key,
		info: *info,
	}
	c.providers[p.name] = p
}

func GetProviderInfo(ev *clientv3.Event) *ProviderInfo {
	info := &ProviderInfo{}
	err := json.Unmarshal([]byte(ev.Kv.Value), info)
	if err != nil {
		log.Fatal(err)
	}
	return info
}

func (c *Consumer) WatchProvider() {
	chanWatch := c.client.Watch(context.Background(), c.path, clientv3.WithPrefix())
	for wresp := range chanWatch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				info := GetProviderInfo(ev)
				log.Println(string(ev.Kv.Key) + " " + info.IP + "is Connecting!")
				c.AddProvider(string(ev.Kv.Key), info)
			case clientv3.EventTypeDelete:
				log.Println(string(ev.Kv.Key) + "Has Been removed!")
				delete(c.providers, string(ev.Kv.Key))
			}
		}
	}

}
