package mbqt

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/grid-x/modbus"
	"github.com/spf13/cobra"
)

func init() {
	mboserial.Flags().StringVar(&serial_dev, "serial-dev", "/dev/ttyUSB0", "Serial device file (it can also be a named pipe).")
	mboserial.Flags().StringVar(&parity, "serial-parity", "N", "Serial parity. One of [N, E, O].")
	mboserial.Flags().IntVar(&baud_rate, "baud-rate", 9600, "Baud Rate in bps.")
	mboserial.Flags().IntVar(&data_bits, "data-bits", 8, "Data bits")
	mboserial.Flags().IntVar(&stop_bits, "stop-bits", 1, "Stop bits")
	mboserial.Flags().IntVar(&timeout, "timeout", 1, "Timeout in s")
}

var (
	serial_dev string
	parity     string
	baud_rate  int
	data_bits  int
	stop_bits  int
	timeout    int

	mboserial = &cobra.Command{
		Use:   "serial <slave-id> <op> <address> [data]",
		Short: "Query a ModBus salve over serial transport.",
		Long: "The command MUST be provided the slave ID of the device it is to query.\n" +
			"An operation should also be specified: these are either 'read' or 'write'\n" +
			"The command also requires the address of the ModBus register to read\n" +
			"The address MUST be provided as a hexadecimal integer. By default, we\n" +
			"assume ModBus registers to be 16-bit (i.e. 2 bytes)" +
			"If writing data to the slave, the data to write should be provided as a hex value\n",
		Args: arg_validation,
		Run: func(cmd *cobra.Command, args []string) {
			handler := modbus.NewRTUClientHandler(serial_dev)

			// Argument already validated on mboserial.Args!
			sid, _ := strconv.Atoi(args[0])

			handler.Parity = parity
			handler.SlaveID = byte(sid)
			handler.BaudRate = baud_rate
			handler.DataBits = data_bits
			handler.StopBits = stop_bits
			handler.Timeout = time.Duration(timeout) * time.Second

			if err := handler.Connect(); err != nil {
				fmt.Printf("Couldn't connect to the ModBus slave @ %s: %v\n", serial_dev, err)
				os.Exit(-1)
			}
			defer handler.Close()

			mb_comms(args, modbus.NewClient(handler))
		},
	}
)
