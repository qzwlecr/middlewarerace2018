package main

import (
	"consumer"
	"provider"
)

func main() {
	go func() {
		provider.NewProvider(
			[]string{"127.0.0.1:2379"},
			"/provider/qaq",
			provider.ProviderInfo{
				IP:     "127.0.0.1",
				CPU:    102400,
				Memory: 409600,
			},
		)
	}()
	go func() {
		 consumer.NewConsumer([]string{"127.0.0.1:2379"}, "/provider")
	}()
	for {
		;
	}

}
