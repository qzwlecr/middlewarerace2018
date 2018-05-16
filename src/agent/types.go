package agent

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"time"
)

const (
	dialTimeout = 5 * time.Second
	MinTTL      = 5
)

type Provider struct {
	name     string
	etcdAddr string
	info     ProviderInfo
	stop     chan error
	leaseId  etcdv3.LeaseID
	client   *etcdv3.Client
}

type Consumer struct {
	path      string
	etcdAddr  string
	providers map[string]*Provider
	client    *etcdv3.Client
}

type ProviderInfo struct {
	IP     string
	CPU    int
	Memory int
}
