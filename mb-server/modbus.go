package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goburrow/serial"
	mbserver "github.com/wilkingj/GoModbusServer"
)

func ServeModbus() (*mbserver.Server, error) {
	serv, err := mbserver.NewServer(uint8(slave_id))
	if err != nil {
		return nil, err
	}

	log.Printf("ModBusServer: instantiated the server\n")

	serv.HoldingRegisters = make([]byte, 653356*2)

	// Fill in some dummy values to make checking stuff easier
	serv.HoldingRegisters[0] = 0xA
	serv.HoldingRegisters[1] = 0xB
	serv.HoldingRegisters[2] = 0xC
	serv.HoldingRegisters[3] = 0xD

	log.Printf("ModBusServer: initialised register addresses 0 and 1\n")

	if err := serv.ListenTCP(fmt.Sprintf("%s:%d", mb_bind_addr, mb_bind_port)); err != nil {
		return nil, err
	}
	log.Printf("ModBusServer: began listening on %s:%d [TCP]\n", mb_bind_addr, mb_bind_port)

	if strings.ToLower(serial_dev) != "none" {
		if err := serv.ListenRTU(
			&serial.Config{
				Address:  serial_dev,
				BaudRate: baud_rate,
				DataBits: data_bits,
				StopBits: stop_bits,
				Parity:   parity,
				Timeout:  time.Duration(timeout) * time.Second,
			}); err != nil {
			return nil, err
		}
		log.Printf("ModBusServer: began listening on %s [Serial]\n", serial_dev)
	}
	return serv, nil
}
