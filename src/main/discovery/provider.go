package discovery

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"encoding/json"
	"context"
	"errors"
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

func (p *Provider) Start() error {

	ch, err := p.keepAlive()
	if err != nil {
		log.Fatal(err)
		return err
	}

	for {
		select {
		case err := <-p.stop:
			p.revoke()
			return err
		case <-p.client.Ctx().Done():
			return errors.New("provider closed")
		case _, ok := <-ch:
			if !ok {
				log.Println("KA channel closed")
				p.revoke()
				return nil
			}
		}
	}
}

func (p *Provider) Stop() {
	p.stop <- nil
}

func (p *Provider) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {

	info := &p.info

	key := "services/" + p.name
	value, _ := json.Marshal(info)

	resp, err := p.client.Grant(context.TODO(), MinTTL)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	_, err = p.client.Put(context.TODO(), key, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	p.leaseId = resp.ID

	return p.client.KeepAlive(context.TODO(), resp.ID)
}

func (p *Provider) revoke() error {

	_, err := p.client.Revoke(context.TODO(), p.leaseId)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Provider:%s stop\n", p.name)
	return err
}
