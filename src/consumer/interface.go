package consumer

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"time"
)

const dialTimeout = 5 * time.Second

type Provider struct {
	name     string
	etcdAddr []string
	delay    uint64
	info     ProviderInfo
	leaseId  etcdv3.LeaseID
	client   *etcdv3.Client
}

type Consumer struct {
	path      string
	etcdAddr  []string
	providers map[string]*Provider
	client    *etcdv3.Client
}

type ProviderInfo struct {
	IP     string
	CPU    int
	Memory int
}
