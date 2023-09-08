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

type DeviceSpec struct {
	RemoteSlaveId    byte
	RemoteModbusAddr uint16
	LocalModbusAddr  uint16
}

var (
	// General settings
	requestWait          int
	localModbusDeviceMap string
	idToModbusAddress    map[string]DeviceSpec = map[string]DeviceSpec{}

	// Local Modbus TCP client
	localModbusBindAddr string
	localModbusBindPort int
	localModbusSlaveId  byte

	// Remote Modbus RTU client over LoRa
	remoteModbusEnable    bool
	remoteModbusSerialDev string
	remoteModbusParity    string
	remoteModbusBaudRate  int
	remoteModbusDataBits  int
	remoteModbusStopBits  int
	remoteModbusTimeout   int

	// UDP data insertion
	udpEnable   bool
	udpBindAddr string
	udpBindPort int

	rootCmd = &cobra.Command{
		Use:   "mb-master",
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
			log.SetFlags((log.Flags() | log.Lshortfile) &^ (log.Ldate | log.Ltime))

			for _, devMapping := range strings.Split(localModbusDeviceMap, ",") {
				mapData := strings.Split(devMapping, ":")
				if len(mapData) != 4 {
					log.Fatalf("error parsing device map: each mapping should have two elements!\n")
				}

				remoteSlaveId, err := strconv.Atoi(mapData[1])
				if err != nil {
					log.Fatalf("error parsing device map: %v\n", err)
				}

				remoteModbusAddr, err := strconv.Atoi(mapData[2])
				if err != nil {
					log.Fatalf("error parsing device map: %v\n", err)
				}

				localModbusAddr, err := strconv.Atoi(mapData[3])
				if err != nil {
					log.Fatalf("error parsing device map: %v\n", err)
				}

				idToModbusAddress[mapData[0]] = DeviceSpec{
					byte(remoteSlaveId), uint16(remoteModbusAddr), uint16(localModbusAddr)}
			}

			log.Printf("parsed device map: %v\n", idToModbusAddress)

			mbServer, err := ServeModbus()
			if err != nil {
				log.Fatalf("couldn't open start the Modbus server: %v\n", err)
			}
			defer mbServer.Close()

			if udpEnable {
				go InsertDataUDP()
			}
			if remoteModbusEnable {
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
	rootCmd.Flags().StringVar(&localModbusDeviceMap, "device-map", "geonosis:2:2:4,corellia:3:2:5", "Mapping of Device IDs -> ModBus Addresses.")
	rootCmd.Flags().IntVar(&requestWait, "request-wait", 60, "Wait time in seconds before the next remote request.")

	// Local ModBus over TCP
	rootCmd.Flags().StringVar(&localModbusBindAddr, "local-mb-bind-address", "127.0.0.1", "Address the ModBus server is to listen on.")
	rootCmd.Flags().IntVar(&localModbusBindPort, "local-mb-bind-port", 1502, "Port the ModBus server is to listen on.")
	rootCmd.Flags().Uint8Var(&localModbusSlaveId, "local-mb-salve-id", 1, "Slave ID for the server to listen on.")

	// Remote ModBus over RTU/Serial
	rootCmd.Flags().BoolVar(&remoteModbusEnable, "remote-mb-enable", true, "Whether to enable data reception over ModBus.")
	rootCmd.Flags().StringVar(&remoteModbusSerialDev, "remote-mb-serial-dev", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'.")
	rootCmd.Flags().StringVar(&remoteModbusParity, "remote-mb-parity", "N", "Serial parity: N | E | O.")
	rootCmd.Flags().IntVar(&remoteModbusBaudRate, "remote-mb-baud-rate", 9600, "Server baud rate in bps.")
	rootCmd.Flags().IntVar(&remoteModbusDataBits, "remote-mb-data-bits", 8, "Data bits.")
	rootCmd.Flags().IntVar(&remoteModbusStopBits, "remote-mb-stop-bits", 1, "Stop bits.")
	rootCmd.Flags().IntVar(&remoteModbusTimeout, "remote-mb-timeout", 60, "Timeout in seconds.")

	// Data input over UDP
	rootCmd.Flags().BoolVar(&udpEnable, "udp-enable", true, "Whether to enable data reception over UDP.")
	rootCmd.Flags().StringVar(&udpBindAddr, "udp-bind-address", "127.0.0.1", "Address to listen on for incoming data over UDP.")
	rootCmd.Flags().IntVar(&udpBindPort, "udp-bind-port", 1503, "Port to listen on for incoming data over UDP.")
}
