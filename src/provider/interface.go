package provider

import (
	"time"

	etcdv3 "github.com/coreos/etcd/clientv3"
)

const (
	dialTimeout = 5 * time.Second
	MinTTL      = 5
)

type Provider struct {
	name        string
	etcdAddr    []string
	info        ProviderInfo
	chanStop    chan error
	leaseId     etcdv3.LeaseID
	client      *etcdv3.Client
	createdConn int
	connIn      []chan []byte
	connOut     []chan []byte
}

type ProviderInfo struct {
	IP     string
	Weight uint32
}
