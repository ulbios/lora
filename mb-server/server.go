package main

import (
	"log"
	"os"
	"time"

	"github.com/goburrow/serial"
	mbserver "github.com/wilkingj/GoModbusServer"
)

func main() {
	serv, err := mbserver.NewServer(1)
	if err != nil {
		log.Fatalf("Couldn't instantiate a ModBus server: %v\n", err)
	}

	serv.HoldingRegisters = make([]byte, 653356*2)
	serv.HoldingRegisters[0] = 0xF
	serv.HoldingRegisters[1] = 0xA
	serv.HoldingRegisters[2] = 0xB

	err = serv.ListenTCP("127.0.0.1:1502")
	if err != nil {
		log.Fatalf("Coulnd't listen on port 1502: %v\n", err)
	}
	defer serv.Close()

	err = serv.ListenRTU(&serial.Config{
		Address:  os.Args[1],
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second})
	if err != nil {
		log.Fatalf("failed to listen, got %v\n", err)
	}

	log.Printf("Began listening on 127.0.0.1:1502!\n")
	for {
		time.Sleep(1 * time.Second)
	}
}
