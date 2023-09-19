package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

func InsertDataLoRa() error {
	localHandler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", localModbusBindAddr, localModbusBindPort))
	if err := localHandler.Connect(); err != nil {
		return err
	}
	defer localHandler.Close()
	log.Printf("instantiated local TCP Modbus handler\n")

	localClient := modbus.NewClient(localHandler)
	log.Printf("instantiated local TCP Modbus client\n")

	remoteClients := map[string]modbus.Client{}
	for id, devInfo := range idToModbusAddress {
		remoteHandler := modbus.NewRTUClientHandler(remoteModbusSerialDev)

		remoteHandler.Parity = remoteModbusParity
		remoteHandler.SlaveId = devInfo.RemoteSlaveId
		remoteHandler.BaudRate = remoteModbusBaudRate
		remoteHandler.DataBits = remoteModbusDataBits
		remoteHandler.StopBits = remoteModbusStopBits
		remoteHandler.Timeout = time.Duration(remoteModbusTimeout) * time.Second

		if err := remoteHandler.Connect(); err != nil {
			log.Fatalf("couldn't connect to the ModBus slave @ %s: %v\n", remoteModbusSerialDev, err)
		}
		defer remoteHandler.Close()
		log.Printf("instantiated remote Modbus RTU handler for %s\n", id)

		remoteClients[id] = modbus.NewClient(remoteHandler)
		log.Printf("instantiated local RTU Modbus client for %s\n", id)
	}

	log.Printf("instantiated all the remote clients!\n")

	for {
		for id, modbusAddresses := range idToModbusAddress {
			log.Printf("requesting data for device %s...\n", id)
			remoteData, err := ReadHoldingRegister(remoteClients[id], modbusAddresses.RemoteModbusAddr)
			if err != nil {
				log.Printf("error receiving data: %v\n", err)
			} else {
				log.Printf("inserting received data '%d' into local server @ %#x\n", remoteData, modbusAddresses.LocalModbusAddr)
				_, err = localClient.WriteSingleRegister(modbusAddresses.LocalModbusAddr, remoteData)
				if err != nil {
					log.Printf("error updating the local Modbus server: %v\n", err)
				}
			}

			log.Printf("waiting for %d seconds until the next request...\n", requestWait)
			time.Sleep(time.Duration(requestWait) * time.Second)
		}
	}
}

func ReadHoldingRegister(c modbus.Client, addr uint16) (uint16, error) {
	rawData, err := c.ReadHoldingRegisters(addr, 1)
	if err != nil {
		return 0, err
	}
	return uint16(rawData[0])<<8 | uint16(rawData[1]), nil
}
