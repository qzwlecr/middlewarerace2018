package discovery

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"encoding/json"
	"context"
)

func NewProvider(name string, info ProviderInfo, endpoints []string) *Provider {

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	cli, err := clientv3.New(cfg)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	p := &Provider{
		name:   name,
		info:   info,
		stop:   make(chan error),
		client: cli,
	}

	return p

}

func (p *Provider) Start() {
	ch := p.keepAlive()

	for {
		select {
		case <-p.stop:
			p.revoke()
			return
		case <-p.client.Ctx().Done():
			return
		case _, ok := <-ch:
			if !ok {
				log.Println("KA channel closed")
				p.revoke()
				return
			}
		}
	}
}

func (p *Provider) Stop() {
	p.stop <- nil
}

func (p *Provider) keepAlive() <-chan *clientv3.LeaseKeepAliveResponse {

	info := &p.info

	key := p.name
	value, _ := json.Marshal(info)

	resp, err := p.client.Grant(context.Background(), MinTTL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	_, err = p.client.Put(context.Background(), key, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	p.leaseId = resp.ID

	ret, err := p.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return ret
}

func (p *Provider) revoke() error {

	_, err := p.client.Revoke(context.Background(), p.leaseId)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
