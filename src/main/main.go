package main

import (
	"provider"
	"flag"
	"consumer"
	"net"
	"os"
	"log"
)

func main() {
	logsDir := flag.String("l", "/root/logs", "")
	etcdUrl := flag.String("u", "http://etcd:2379", "")
	memory := flag.Int("m", 1536, "")
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
		ip, _ := net.InterfaceAddrs()
		log.Println("Provider's ip is:", ip)
		provider.NewProvider(
			[]string{*etcdUrl},
			"/provider/" + *name,
			provider.ProviderInfo{
				//TODO
				IP:     ip[0].String(),
				Memory: *memory,
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
