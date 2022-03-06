package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goburrow/serial"
	mbserver "github.com/wilkingj/GoModbusServer"
)

var (
	bind_addr string
	bind_port int

	serial_dev string
	parity     string
	slave_id   int
	baud_rate  int
	data_bits  int
	stop_bits  int
	timeout    int
)

func main() {
	// TCP/IP stuff
	flag.StringVar(&bind_addr, "bind-address", "127.0.0.1", "Address to listen on")
	flag.IntVar(&bind_port, "bind-port", 1502, "Port to listen on")

	// Serial stuff
	flag.StringVar(&serial_dev, "serial-device", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'")
	flag.StringVar(&parity, "serial-parity", "N", "Serial parity: N | E | O")
	flag.IntVar(&slave_id, "slave-id", 1, "Slave ID to run as [1, 247]")
	flag.IntVar(&baud_rate, "baud-rate", 115200, "Server baud rate in bps")
	flag.IntVar(&data_bits, "data-bits", 8, "Data bits")
	flag.IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	flag.IntVar(&timeout, "timeout", 10, "Timeout in s")

	flag.Parse()

	serv, err := mbserver.NewServer(uint8(slave_id))
	if err != nil {
		log.Fatalf("Couldn't instantiate the ModBus server: %v\n", err)
	}

	serv.HoldingRegisters = make([]byte, 653356*2)

	// Fill in some dummy values to make checking stuff easier
	serv.HoldingRegisters[0] = 0xA
	serv.HoldingRegisters[1] = 0xB
	serv.HoldingRegisters[2] = 0xC
	serv.HoldingRegisters[3] = 0xD

	if err := serv.ListenTCP(fmt.Sprintf("%s:%d", bind_addr, bind_port)); err != nil {
		log.Fatalf("Couldn't listen on %s:%d -> %v\n", bind_addr, bind_port, err)
	}
	defer serv.Close()
	log.Printf("Began listening on %s:%d\n", bind_addr, bind_port)

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
			log.Fatalf("Failed to listen on %s: %v\n", serial_dev, err)
		}
		log.Printf("Began listening on %s\n", serial_dev)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
