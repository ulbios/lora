package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
	"ulbios/rfm9x-driver"

	"github.com/grid-x/modbus"

	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

type Data_point struct {
	Msg  string `json:"msg"`
	Data int    `json:"data"`
}

var param_to_addr map[string]uint16 = map[string]uint16{
	"v_1": 0, "v_2": 1, "c_1": 2, "c_2": 3,
}

var (
	poll_interval int
	read_param    string

	serial_dev string
	parity     string
	slave_id   int
	baud_rate  int
	data_bits  int
	stop_bits  int
	timeout    int

	lora_enable   bool
	lora_spi_port string
)

func get_mb_cli(serial_dev string) (modbus.Client, *modbus.RTUClientHandler) {
	handler := modbus.NewRTUClientHandler(serial_dev)

	handler.Parity = parity
	handler.SlaveID = byte(slave_id)
	handler.BaudRate = baud_rate
	handler.DataBits = data_bits
	handler.StopBits = stop_bits
	handler.Timeout = time.Duration(timeout) * time.Second

	if err := handler.Connect(); err != nil {
		log.Fatalf("couldn't connect to the ModBus slave @ %s: %v\n", serial_dev, err)
		return nil, nil
	}

	return modbus.NewClient(handler), handler
}

func read_420(c modbus.Client) (uint32, error) {
	r_data, err := c.ReadHoldingRegisters(param_to_addr[read_param], 1)
	if err != nil {
		return 0, err
	}
	return uint32(r_data[0])<<8 | uint32(r_data[1]), nil
}

func get_lora_cli() (*rfm9x.Dev, spi.PortCloser) {
	if _, err := host.Init(); err != nil {
		log.Printf("error initialising Periph: %v\n", err)
		return nil, nil
	}

	p, err := spireg.Open(lora_spi_port)
	if err != nil {
		log.Printf("error opening the SPI port: %v\n", err)
		return nil, nil
	}

	d_opts := rfm9x.DefaultOpts

	radio, err := rfm9x.New(
		p,
		&d_opts,
	)
	if err != nil {
		log.Fatalf("error instantiating the LoRa radio: %v\n", err)
		return nil, nil
	}

	return radio, p
}

func send_over_lora(r *rfm9x.Dev, msg string, data int) error {
	enc_payload, err := json.Marshal(Data_point{Msg: msg, Data: data})
	if err != nil {
		log.Printf("error marshalling data: %v\n", err)
		return err
	}
	return r.Send(enc_payload)
}

func main() {
	// General stuff
	flag.IntVar(&poll_interval, "poll-interval", 30, "Poll interval in seconds")
	flag.StringVar(&read_param, "read-param", "c-1", "Parameter to read over ModBus")

	// Serial stuff
	flag.StringVar(&serial_dev, "serial-device", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'")
	flag.StringVar(&parity, "serial-parity", "N", "Serial parity: N | E | O")
	flag.IntVar(&slave_id, "slave-id", 1, "Slave ID to connect to [1, 247]")
	flag.IntVar(&baud_rate, "baud-rate", 115200, "Server baud rate in bps")
	flag.IntVar(&data_bits, "data-bits", 8, "Data bits")
	flag.IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	flag.IntVar(&timeout, "timeout", 10, "Timeout in s")

	// LoRa stuff
	flag.BoolVar(&lora_enable, "lora-enable", false, "Whether to enable data reception over LoRa")
	flag.StringVar(&lora_spi_port, "lora-spi-port", "/dev/spidev0.1", "SPI address the radio is on")

	flag.Parse()

	// Remove leading date and time from log messages
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	mb_cli, mb_handler := get_mb_cli(serial_dev)
	defer mb_handler.Close()

	lora_cli, lora_pc := get_lora_cli()
	defer lora_pc.Close()

	for range time.Tick(time.Second * time.Duration(poll_interval)) {
		data, err := read_420(mb_cli)
		if err != nil {
			log.Printf("error reading 420 data: %v\n", err)
			continue
		}
		if err := send_over_lora(lora_cli, func() string { hn, _ := os.Hostname(); return hn }(), int(data)); err != nil {
			log.Printf("error sending data over LoRa: %v\n", err)
		}
	}
}
