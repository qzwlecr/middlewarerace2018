package main

import (
	"consumer"
	"flag"
	"log"
	"net"
	//_ "net/http/pprof"
	"os"
	"provider"
)

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	logsDir := flag.String("l", "/root/logs", "")
	etcdUrl := flag.String("u", "http://etcd:2379", "")
	weight := flag.Uint("w", 1, "")
	types := flag.String("t", "provider", "")
	name := flag.String("n", "small", "")
	flag.Parse()
	f, err := os.OpenFile(*logsDir+"/own.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Start!" + *types + "-" + *name + "!")
	if *types == "provider" {
		provider.NewProvider(
			[]string{*etcdUrl},
			"/provider/"+*name,
			provider.ProviderInfo{
				IP:     GetLocalIP(),
				Weight: uint32(*weight),
			},
		)

	} else {
		consumer.NewConsumer(
			[]string{*etcdUrl},
			"/provider",
		)

	}
	<-(chan int)(nil)
}
