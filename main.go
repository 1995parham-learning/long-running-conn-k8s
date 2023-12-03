package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// Configure NATS client options
	opts := []nats.Option{
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to %v", nc.ConnectedUrl())
		}),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("Disconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Fatalf("Connection closed. Reason: %v", nc.LastError())
		}),
	}

	// Connect to the NATS server
	nc, err := nats.Connect("nats://172.19.1.4:4222", nats.PingInterval(5*time.Second), nats.MaxPingsOutstanding(2), opts[0], opts[1], opts[2])
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Publish a message every second
	for {
		if nc.IsConnected() {
			if err := nc.Publish("subject", []byte("hello")); err != nil {
				log.Printf("Error publishing message: %v", err)
			} else {
				log.Println("Published message: hello")
			}
		} else {
			log.Println("Not connected. Waiting to reestablish the connection.")
		}
		time.Sleep(1 * time.Second)
	}
}
