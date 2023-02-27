package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"ulbios/rfm9x-driver"

	mbclient "github.com/goburrow/modbus"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

var lora_debug = []rfm9x.Log_level{
	rfm9x.LogLevelErr,
	rfm9x.LogLevelWarn,
	rfm9x.LogLevelInfo,
	rfm9x.LogLevelDebug,
	rfm9x.LogLevelRegIO,
}

func InsertDataLoRa(freq int64) error {
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
	d_opts.FrequencyMHz = freq
	d_opts.LogLevel = lora_debug[lora_debug_level]

	radio, err := rfm9x.New(
		p,
		&d_opts,
	)
	if err != nil {
		log.Fatalf("Error opening the SPI device: %v", err)
	}

	var dp DataPoint

	for {
		enc_pkt, err := radio.Receive(
			time.Duration(lora_recv_wait)*time.Millisecond, time.Duration(lora_recv_timeout)*time.Millisecond)
		if err != nil {
			log.Printf("LoRa: error receiving data: %v\n", err)
			continue
		}
		if len(enc_pkt) < 5 {
			log.Printf("LoRa: the received packet is too short: %v (len %d)\n", enc_pkt, len(enc_pkt))
			continue
		}
		if err := json.Unmarshal(enc_pkt[4:], &dp); err != nil {
			log.Printf("LoRa: error unmarshalling data: %v [%s]\n", err, enc_pkt[4:])
			continue
		}

		log.Printf("LoRa: received -> %#v\n", dp)

		addr, ok := id_to_mb_addr[dp.Id]
		if !ok {
			log.Printf("LoRa: received a nonexistent ID: %s\n", dp.Id)
			continue
		}

		if dp.Data == 0 {
			log.Printf("LoRa: ignoring incorrect data value %d and sticking to the last one\n", dp.Data)
			continue
		}

		_, err = client.WriteSingleRegister(addr, uint16(dp.Data))
		if err != nil {
			log.Printf("LoRa: error updating ModBus server: %v\n", err)
			continue
		}

		log.Printf("LoRa: sent data to ModBus server @ %d\n", addr)
	}
}
