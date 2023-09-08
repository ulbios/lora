package main

import (
	"fmt"
	"log"

	mbserver "github.com/wilkingj/GoModbusServer"
)

const MODBUS_SERVER_BUFFSIZE int = 100

func ServeModbus() (*mbserver.Server, error) {
	serv, err := mbserver.NewServer(localModbusSlaveId)
	if err != nil {
		return nil, err
	}

	log.Printf("instantiated the server\n")

	serv.HoldingRegisters = make([]byte, MODBUS_SERVER_BUFFSIZE*2)

	// Fill in some dummy values to make checking stuff easier
	serv.HoldingRegisters[0] = 0xAB
	serv.HoldingRegisters[1] = 0xCD

	log.Printf("initialised register addresses 0 and 1 with 0xABCD\n")

	if err := serv.ListenTCP(fmt.Sprintf("%s:%d", localModbusBindAddr, localModbusBindPort)); err != nil {
		return nil, err
	}
	log.Printf("began listening on %s:%d [TCP]\n", localModbusBindAddr, localModbusBindPort)

	return serv, nil
}
