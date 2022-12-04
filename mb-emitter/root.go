package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-co-op/gocron"
)

type DataPoint struct {
	Id   string `json:"id"`
	Data int    `json:"data"`
}

var (
	soc      string
	sysfsPin int

	poll_interval string
	read_param    string

	serial_dev string
	parity     string
	slave_id   int
	baud_rate  int
	data_bits  int
	stop_bits  int
	timeout    int

	lora_enable       bool
	lora_spi_port     string
	carrier_frequency int64

	param_to_addr map[string]uint16 = map[string]uint16{
		"v_1": 0, "v_2": 1, "c_1": 2, "c_2": 3,
	}

	rootCmd = &cobra.Command{
		Use:   "mb-emitter",
		Short: "A ModBus Emitter reading 4-20 mA data and sending it over LoRa.",
		Long: "This executable implements a ModBus client over RTU-over-serial that queries a 4-20 -- ModBus transducer.\n" +
			"Retrieved data is then sent over LoRa so that it can be acquired by a main gateway.\n",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("no arguments should be provided (configuration is done through flags)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Remove leading date and time from log messages
			log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

			mb_cli, mb_handler := GetModBusCli(serial_dev)
			defer mb_handler.Close()

			lora_cli, lora_pc, err := GetLoRaCli(carrier_frequency)
			if err != nil {
				log.Fatalf("error opening the LoRa radio: %v\n", err)
			}
			defer lora_pc.Close()

			hn, _ := os.Hostname()

			s := gocron.NewScheduler(time.UTC)

			_, _ = s.CronWithSeconds(poll_interval).Do(func() {
				data, err := Read420(mb_cli)
				if err != nil {
					log.Printf("error reading 420 data: %v\n", err)
					return
				}

				if err := SendOverLoRa(lora_cli, hn, int(data)); err != nil {
					log.Printf("error sending data over LoRa: %v\n", err)
				}
			})
			s.StartBlocking()
		},
	}
)

func init() {
	// Disable Cobra completions
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Hardware selection
	rootCmd.Flags().StringVar(&soc, "soc-model", "rpi", "The SoC to run on.")
	rootCmd.Flags().IntVar(&sysfsPin, "sysfs-pin", 21, "GPIO pin to control through SysFs")

	// The Cron syntax follows https://en.wikipedia.org/wiki/Cron
	rootCmd.Flags().StringVar(&poll_interval, "poll-interval", "0 * * * * *",
		"Poll interval as a Cron expression [1 minute @ minute start by default]")
	rootCmd.Flags().StringVar(&read_param, "read-param", "c-1", "Parameter to read over ModBus")

	// ModBus over RTU/Serial
	rootCmd.Flags().StringVar(&serial_dev, "serial-device", "/dev/ttyUSB0", "Serial device to listen on: a path to a device or 'NONE'")
	rootCmd.Flags().StringVar(&parity, "serial-parity", "N", "Serial parity: N | E | O")
	rootCmd.Flags().IntVar(&slave_id, "slave-id", 1, "Slave ID to connect to [1, 247]")
	rootCmd.Flags().IntVar(&baud_rate, "baud-rate", 9600, "Server baud rate in bps")
	rootCmd.Flags().IntVar(&data_bits, "data-bits", 8, "Data bits")
	rootCmd.Flags().IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	rootCmd.Flags().IntVar(&timeout, "timeout", 5, "Timeout in s")

	// Data output over LoRa
	rootCmd.Flags().BoolVar(&lora_enable, "lora-enable", false, "Whether to enable data reception over LoRa")
	rootCmd.Flags().StringVar(&lora_spi_port, "lora-spi-port", "/dev/spidev0.1", "SPI address the radio is on")
	rootCmd.Flags().Int64Var(&carrier_frequency, "lora-freq", 868, "Carrier frequency in MHz")
}
