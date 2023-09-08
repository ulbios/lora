package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	mbclient "github.com/goburrow/modbus"
)

type UDPDataPoint struct {
	DeviceId string
	Data     uint16
}

func InsertDataUDP() error {
	handler := mbclient.NewTCPClientHandler(fmt.Sprintf("%s:%d", localModbusBindAddr, localModbusBindPort))
	if err := handler.Connect(); err != nil {
		return err
	}
	defer handler.Close()

	log.Printf("instantiated ModBus handler\n")

	client := mbclient.NewClient(handler)

	log.Printf("instantiated ModBus client\n")

	inSck, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%d", udpBindAddr, udpBindPort))
	if err != nil {
		return err
	}
	defer inSck.Close()

	log.Printf("began listening on %s:%d over UDP\n", udpBindAddr, udpBindPort)

	uBuff := make([]byte, 1024)
	var dp UDPDataPoint

	for {
		n, _, _ := inSck.ReadFrom(uBuff)
		if err := json.Unmarshal(uBuff[:n], &dp); err != nil {
			log.Printf("UDP: error decoding JSON: %v\n", err)
			continue
		}

		log.Printf("received: %#v\n", dp)

		addresses, ok := idToModbusAddress[dp.DeviceId]
		if !ok {
			log.Printf("received a nonexistent device ID: %s\n", dp.DeviceId)
			continue
		}

		_, err := client.WriteSingleRegister(addresses.LocalModbusAddr, dp.Data)
		if err != nil {
			log.Printf("error updating the local ModBus server: %v\n", err)
			continue
		}

		log.Printf("sent data to ModBus server for device ID %q @ %#x\n",
			dp.DeviceId, addresses.LocalModbusAddr)
	}
}
