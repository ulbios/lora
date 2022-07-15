package main

import (
	"encoding/json"
	"log"
	"time"
	"ulbios/rfm9x-driver"

	"github.com/grid-x/modbus"

	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/sysfs"
)

func GetModBusCli(serial_dev string) (modbus.Client, *modbus.RTUClientHandler) {
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

	log.Printf("Connected to ModBus slave @ %s with Slave ID %d\n", serial_dev, slave_id)

	return modbus.NewClient(handler), handler
}

func Read420(c modbus.Client) (uint32, error) {
	log.Printf("Trying to read address %d on %s\n", param_to_addr[read_param], serial_dev)
	r_data, err := c.ReadHoldingRegisters(param_to_addr[read_param], 1)
	if err != nil {
		return 0, err
	}
	return uint32(r_data[0])<<8 | uint32(r_data[1]), nil
}

func GetLoRaCli() (*rfm9x.Dev, spi.PortCloser, error) {
	if _, err := host.Init(); err != nil {
		log.Printf("error initialising Periph: %v\n", err)
		return nil, nil, err
	}

	p, err := spireg.Open(lora_spi_port)
	if err != nil {
		log.Printf("error opening the SPI port: %v\n", err)
		return nil, nil, err
	}

	d_opts := rfm9x.DefaultOpts

	if soc == "opi" {
		d_opts.ResetPin = sysfs.Pins[sysfsPin]
	}

	radio, err := rfm9x.New(
		p,
		&d_opts,
	)
	if err != nil {
		log.Fatalf("error instantiating the LoRa radio: %v\n", err)
		return nil, nil, err
	}
	return radio, p, nil
}

func SendOverLoRa(r *rfm9x.Dev, id string, data int) error {
	dp := DataPoint{Id: id, Data: data}

	log.Printf("sending %#v", dp)

	enc_payload, err := json.Marshal(dp)
	if err != nil {
		log.Printf("error marshalling data: %v\n", err)
		return err
	}

	return r.Send(enc_payload)
}
