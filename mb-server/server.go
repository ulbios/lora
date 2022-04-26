package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"ulbios/rfm9x-driver"

	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"

	mbclient "github.com/goburrow/modbus"
	"github.com/goburrow/serial"
	mbserver "github.com/wilkingj/GoModbusServer"
)

type Data_point struct {
	Msg  string `json:"msg"`
	Data int    `json:"data"`
}

var (
	mb_bind_addr string
	mb_bind_port int

	serial_dev string
	parity     string
	slave_id   int
	baud_rate  int
	data_bits  int
	stop_bits  int
	timeout    int

	udp_enable    bool
	udp_bind_addr string
	udp_bind_port int

	lora_enable   bool
	lora_spi_port string
)

func serve_mb() (*mbserver.Server, error) {
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

func insert_data_lora() error {
	handler := mbclient.NewTCPClientHandler(fmt.Sprintf("%s:%d", mb_bind_addr, mb_bind_port))
	if err := handler.Connect(); err != nil {
		return err
	}
	defer handler.Close()

	log.Printf("LoRa: instantiated ModBus handler\n")

	client := mbclient.NewClient(handler)

	log.Printf("LoRa: instantiated ModBus client\n")

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	log.Printf("LoRa: initialised Periph\n")

	p, err := spireg.Open(lora_spi_port)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	fmt.Printf("LoRa: correctly opened SPI port %s\n", p)

	d_opts := rfm9x.DefaultOpts

	radio, err := rfm9x.New(
		p,
		&d_opts,
	)
	if err != nil {
		log.Fatalf("Error opening the SPI device: %v", err)
	}

	var dp Data_point

	for {
		enc_pkt, err := radio.Receive()
		if err != nil {
			log.Printf("LoRa: error receiving data: %v\n", err)
			continue
		}
		if err := json.Unmarshal(enc_pkt[4:], &dp); err != nil {
			log.Fatalf("LoRa: error unmarshalling data: %v\n", err)
			continue
		}

		log.Printf("LoRa: received -> %#v\n", dp)

		_, err = client.WriteSingleRegister(4, uint16(dp.Data))
		if err != nil {
			log.Printf("LoRa: error updating ModBus server: %v\n", err)
			continue
		}

		log.Printf("LoRa: sent data to ModBus server\n")
	}
}

func insert_data_udp() error {
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
	var dp Data_point

	for {
		n, _, _ := in_sck.ReadFrom(ubuff)
		if err := json.Unmarshal(ubuff[:n], &dp); err != nil {
			log.Printf("UDP: error decoding JSON: %v\n", err)
			continue
		}

		log.Printf("UDP: received -> %#v\n", dp)

		_, err := client.WriteSingleRegister(4, uint16(dp.Data))
		if err != nil {
			log.Printf("UDP: error updating ModBus server: %v\n", err)
			continue
		}

		log.Printf("UDP: sent data to ModBus server\n")
	}
}

func main() {
	// ModBus server stuff
	flag.StringVar(&mb_bind_addr, "mb-bind-address", "127.0.0.1", "Address to listen on")
	flag.IntVar(&mb_bind_port, "mb-bind-port", 1502, "Port to listen on")

	// Serial stuff
	flag.StringVar(&serial_dev, "serial-device", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'")
	flag.StringVar(&parity, "serial-parity", "N", "Serial parity: N | E | O")
	flag.IntVar(&slave_id, "slave-id", 1, "Slave ID to run as [1, 247]")
	flag.IntVar(&baud_rate, "baud-rate", 115200, "Server baud rate in bps")
	flag.IntVar(&data_bits, "data-bits", 8, "Data bits")
	flag.IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	flag.IntVar(&timeout, "timeout", 10, "Timeout in s")

	// UDP data input stuff
	flag.BoolVar(&udp_enable, "udp-enable", true, "Whether to enable data reception over UDP")
	flag.StringVar(&udp_bind_addr, "udp-bind-address", "127.0.0.1", "Address to listen on for incoming data over UDP")
	flag.IntVar(&udp_bind_port, "udp-bind-port", 1503, "Port to listen on for incoming data over UDP")

	// LoRa stuff
	flag.BoolVar(&lora_enable, "lora-enable", false, "Whether to enable data reception over LoRa")
	flag.StringVar(&lora_spi_port, "lora-spi-port", "/dev/spidev0.1", "SPI address the radio is on")

	flag.Parse()

	// Remove leading date and time from log messages
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	mb_server, err := serve_mb()
	if err != nil {
		log.Fatalf("main: couldn't open start the modbus server: %v\n", err)
	}
	defer mb_server.Close()

	if udp_enable {
		go insert_data_udp()
	}
	if lora_enable {
		go insert_data_lora()
	}

	sig_ch := make(chan os.Signal, 1)
	signal.Notify(sig_ch, os.Interrupt)

	for range sig_ch {
		log.Printf("main: quitting due to SIGINT!\n")
		return
	}
}
