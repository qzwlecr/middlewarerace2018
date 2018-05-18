package provider

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"log"
	"encoding/json"
	"context"
)

//NewProvider receive etcd server address, the service name, and the service info.
//And return the consumer which has been started.
func NewProvider(endpoints []string, name string, info ProviderInfo) *Provider {

	cfg := etcdv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	cli, err := etcdv3.New(cfg)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	p := &Provider{
		name:   name,
		etcdAddr: endpoints,
		info:   info,
		stop:   make(chan error),
		client: cli,
	}

	go p.Start()

	return p

}

//Start shouldn't be called manually.
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

//Stop must be used for closing connection.
func (p *Provider) Stop() {
	p.stop <- nil
}

//keepAlive receive the etcdv3.response, and update lease.
func (p *Provider) keepAlive() <-chan *etcdv3.LeaseKeepAliveResponse {

	info := &p.info

	key := p.name
	value, _ := json.Marshal(info)

	resp, err := p.client.Grant(context.Background(), MinTTL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	_, err = p.client.Put(context.Background(), key, string(value), etcdv3.WithLease(resp.ID))
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

//revoke is a wrapper of etcd.Revoke.
func (p *Provider) revoke() error {

	_, err := p.client.Revoke(context.Background(), p.leaseId)
	if err != nil {
		log.Fatal(err)
	}
	return err
}