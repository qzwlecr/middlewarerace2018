package consumer

import (
	etcdv3 "github.com/coreos/etcd/clientv3"
	"time"
	"protocol"
	"net"
	"sync"
)

const (
	dialTimeout = 5 * time.Second
	queueSize   = 1024
	connsSize   = 1024
)

type Provider struct {
	name     string
	etcdAddr []string
	delay    uint64
	info     ProviderInfo
	leaseId  etcdv3.LeaseID
	client   *etcdv3.Client
	chanIn   chan protocol.CustRequest
	conns    []Connection
}

type Connection struct {
	conn     net.Conn
	consumer *Consumer
	provider *Provider
}

type Consumer struct {
	path      string
	etcdAddr  []string
	cnvt      protocol.SimpleConverter
	answer    map[uint64]chan []byte
	answerMu  sync.RWMutex
	providers map[string]*Provider
	client    *etcdv3.Client
}

type ProviderInfo struct {
	IP     string
	CPU    int
	Memory int
}
