package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type DataPoint struct {
	Id   string `json:"id"`
	Data int    `json:"data"`
}

var (
	dev_map string

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

	id_to_mb_addr map[string]uint16 = map[string]uint16{}

	rootCmd = &cobra.Command{
		Use:   "mb-server",
		Short: "A ModBus Server capable of receiving data over UDP and LoRa.",
		Long: "This executable implements a ModBus server over both TCP and RTU-over-serial connections.\n" +
			"It's been written to accept incoming data over UDP and LoRa to then make it available through.\n" +
			"The ModBus server for machines positioned 'behind' this one",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("no arguments should be provided (configuration is done through flags)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Remove leading date and time from log messages
			log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

			for _, dev_mapping := range strings.Split(dev_map, ",") {
				map_data := strings.Split(dev_mapping, ":")
				if len(map_data) != 2 {
					log.Fatalf("error parsing device map: each mapping should have two elements!\n")
				}

				mb_addr, err := strconv.Atoi(map_data[1])
				if err != nil {
					log.Fatalf("error parsing device map: %v\n", err)
				}

				id_to_mb_addr[map_data[0]] = uint16(mb_addr)
			}

			mb_server, err := ServeModbus()
			if err != nil {
				log.Fatalf("couldn't open start the modbus server: %v\n", err)
			}
			defer mb_server.Close()

			if udp_enable {
				go InsertDataUDP()
			}
			if lora_enable {
				go InsertDataLoRa()
			}

			sig_ch := make(chan os.Signal, 1)
			signal.Notify(sig_ch, os.Interrupt)

			for range sig_ch {
				return
			}
		},
	}
)

func init() {
	// Disable Cobra completions
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// General
	rootCmd.Flags().StringVar(&dev_map, "device-map", "coruscant:4,geonosis:5", "Mapping of Device IDs -> ModBus Addresses")

	// ModBus over TCP
	rootCmd.Flags().StringVar(&mb_bind_addr, "mb-bind-address", "127.0.0.1", "Address the ModBus server is to listen on")
	rootCmd.Flags().IntVar(&mb_bind_port, "mb-bind-port", 1502, "Port the ModBus server is to listen on")

	// ModBus over RTU/Serial
	rootCmd.Flags().StringVar(&serial_dev, "serial-device", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'")
	rootCmd.Flags().StringVar(&parity, "serial-parity", "N", "Serial parity: N | E | O")
	rootCmd.Flags().IntVar(&slave_id, "slave-id", 1, "Slave ID to run as [1, 247]")
	rootCmd.Flags().IntVar(&baud_rate, "baud-rate", 115200, "Server baud rate in bps")
	rootCmd.Flags().IntVar(&data_bits, "data-bits", 8, "Data bits")
	rootCmd.Flags().IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	rootCmd.Flags().IntVar(&timeout, "timeout", 10, "Timeout in s")

	// Data input over UDP
	rootCmd.Flags().BoolVar(&udp_enable, "udp-enable", true, "Whether to enable data reception over UDP")
	rootCmd.Flags().StringVar(&udp_bind_addr, "udp-bind-address", "127.0.0.1", "Address to listen on for incoming data over UDP")
	rootCmd.Flags().IntVar(&udp_bind_port, "udp-bind-port", 1503, "Port to listen on for incoming data over UDP")

	// Data input over LoRa
	rootCmd.Flags().BoolVar(&lora_enable, "lora-enable", false, "Whether to enable data reception over LoRa")
	rootCmd.Flags().StringVar(&lora_spi_port, "lora-spi-port", "/dev/spidev0.1", "SPI address the radio is on")
}
