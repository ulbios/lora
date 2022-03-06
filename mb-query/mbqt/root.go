package mbqt

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/grid-x/modbus"
	"github.com/spf13/cobra"
)

var (
	// Not providing a func() for the `Run` field forces the program
	// to show a general help message if ran with no sub-commands
	rootCmd = &cobra.Command{
		Use:   "mq-query",
		Short: "A simple client for querying ModBus devices.",
		Long: "The mb-query tool is a simple, standalone executable for querying ModBus devices.\n" +
			"It supports several underlying transport protocols (namely RTU-over-serial-link and TCP/IP).\n",
	}

	ops = map[string]bool{
		"read":  true,
		"write": true,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Disable completion please!
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add the different sub-commands
	rootCmd.AddCommand(mbotcp)
	rootCmd.AddCommand(mboserial)
}

func arg_validation(cmd *cobra.Command, args []string) error {
	if len(args) < 3 || len(args) > 4 {
		return fmt.Errorf("wrong argument count O_o")
	}

	sid, err := strconv.Atoi(args[0])
	if err != nil || sid < 1 || sid > 247 {
		return fmt.Errorf("invalid slave ID %s. It should be larger than 0 and less than 247 [%v]", args[0], err)
	}

	if !ops[strings.ToLower(args[1])] {
		return fmt.Errorf("invalid operation %s", args[1])
	}

	// Overwrite the original to avoid checks later down the road!
	args[1] = strings.ToLower(args[1])

	if args[1] == "write" {
		if len(args) != 4 {
			return fmt.Errorf("data should be provided when writing to the ModBus slave")
		}
		if _, err := strconv.ParseUint(args[3], 16, 16); err != nil {
			return fmt.Errorf("the provided data (i.e. %s) cannot be interpreted as a hex number: %v", args[3], err)
		}
	}

	_, err = strconv.ParseUint(args[2], 16, 16)
	if err != nil {
		return fmt.Errorf("invalid ModBus address %s. It should be a non-negative integer [%v]", args[2], err)
	}

	return nil
}

func mb_comms(args []string, cli modbus.Client) {
	// The input arguments have already been validated through the mbotcp.Args function!
	addr, _ := strconv.ParseUint(args[1], 16, 16)

	switch args[1] {
	case "read":
		rd, err := read_holding_register(cli, uint16(addr))
		if err != nil {
			fmt.Printf("Couldn't read data from the ModBus slave: %v\n", err)
			os.Exit(-1)
		}
		fmt.Printf("Data for address %#x:\n\tDecimal -> %d\n\tHex -> %#x\n\tOctal -> %#o\n", addr, rd, rd, rd)

	case "write":
		data, _ := strconv.ParseUint(args[3], 16, 16)
		if err := write_holding_register(cli, uint16(addr), uint16(data)); err != nil {
			fmt.Printf("Couldn't write the data to the ModBus slave: %v\n", err)
			os.Exit(-1)
		}
		fmt.Printf("Correctly wrote %#x at address %#x on the ModBus slave\n", data, addr)
	}
}

func read_holding_register(c modbus.Client, addr uint16) (uint32, error) {
	raw_data, err := c.ReadHoldingRegisters(addr, 1)
	if err != nil {
		return 0, err
	}
	// log.Printf("Raw data for register @ %#x: %#v\n", addr, raw_data)
	return uint32(raw_data[0])<<8 | uint32(raw_data[1]), nil
}

func write_holding_register(c modbus.Client, addr, data uint16) error {
	_, err := c.WriteSingleRegister(addr, data)
	if err != nil {
		return err
	}
	return nil
}
