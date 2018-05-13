package discovery

import (
	"github.com/coreos/etcd/clientv3"
	"time"
)

const (
	dialTimeout = 5 * time.Second
	MinTTL      = 5
)

type Provider struct {
	name    string
	info    ProviderInfo
	stop    chan error
	leaseId clientv3.LeaseID
	client  *clientv3.Client
}

type Consumer struct {
	providers map[string]*Provider
	path      string
	client    *clientv3.Client
}

type ProviderInfo struct {
	IP     string
	CPU    int
	Memory int
}
