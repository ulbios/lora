package mbqt

import (
	"fmt"
	"os"

	"github.com/grid-x/modbus"
	"github.com/spf13/cobra"
)

func init() {
	mbotcp.Flags().StringVar(&host, "host", "127.0.0.1", "IP address identifying the ModBus slave.")
	mbotcp.Flags().IntVar(&port, "port", 1502, "Port the ModBus slave is listening on.")
}

var (
	host string
	port int

	mbotcp = &cobra.Command{
		Use:   "tcp <slave-id> <op> <address> [data]",
		Short: "Query a ModBus salve over TCP.",
		Long: "The command MUST be provided the slave ID of the device it is to query.\n" +
			"An operation should also be specified: these are either 'read' or 'write'\n" +
			"The command also requires the address of the ModBus register to read.\n" +
			"The address MUST be provided as a hexadecimal integer. By default, we\n" +
			"assume ModBus registers to be 16-bit (i.e. 2 bytes) long." +
			"If writing data to the slave, the data to write should be provided as a hex value\n",
		Args: arg_validation,
		Run: func(cmd *cobra.Command, args []string) {
			handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", host, port))
			if err := handler.Connect(); err != nil {
				fmt.Printf("Couldn't connect to the ModBus slave: %v", err)
				os.Exit(-1)
			}
			defer handler.Close()

			mb_comms(args, modbus.NewClient(handler))
		},
	}
)
