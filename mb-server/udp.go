package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	mbclient "github.com/goburrow/modbus"
)

func InsertDataUDP() error {
	handler := mbclient.NewTCPClientHandler(fmt.Sprintf("%s:%d", mb_bind_addr, mb_bind_port))
	if err := handler.Connect(); err != nil {
		return err
	}
	defer handler.Close()

	log.Printf("UDP: instantiated ModBus handler\n")

	client := mbclient.NewClient(handler)

	log.Printf("UDP: instantiated ModBus client\n")

	in_sck, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%d", udp_bind_addr, udp_bind_port))
	if err != nil {
		return err
	}
	defer in_sck.Close()

	log.Printf("UDP: began listening on %s:%d over UDP\n", udp_bind_addr, udp_bind_port)

	ubuff := make([]byte, 1024)
	var dp DataPoint

	for {
		n, _, _ := in_sck.ReadFrom(ubuff)
		if err := json.Unmarshal(ubuff[:n], &dp); err != nil {
			log.Printf("UDP: error decoding JSON: %v\n", err)
			continue
		}

		log.Printf("UDP: received -> %#v\n", dp)

		addr, ok := id_to_mb_addr[dp.Id]
		if !ok {
			log.Printf("UDP: received a nonexistent ID: %s\n", dp.Id)
			continue
		}

		_, err := client.WriteSingleRegister(addr, uint16(dp.Data))
		if err != nil {
			log.Printf("UDP: error updating ModBus server: %v\n", err)
			continue
		}

		log.Printf("UDP: sent data to ModBus server @ %d\n", addr)
	}
}
