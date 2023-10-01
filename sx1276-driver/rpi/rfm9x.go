package rfm9x

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/host/v3/rpi"
)

// Opts defines configurable options for the device.
type Opts struct {
	// BaudRateMHZ specifies the baudrate at which the SPI
	// 'dialog' happens. This value is assumed to be provided
	// in MegaHertz (i.e. MHz).
	BaudrateMHz uint

	// ResetPin specifies the physical GPIO pin the
	// chip's RESET input is connected to.
	ResetPin gpio.PinIO

	// FrequencyMHz specifies the carrier frequency the radio
	// is to operate at. That is, the frequency that'll be used
	// to send and receive data. This value is assumed to be
	// provided in MegaHertz (i.e. MHz).
	FrequencyMHz int64

	// PreambleLength specifies the size of the preamble to be included
	// on LoRa packets by the radio. This preamble aides in the
	// synchronization of sender and receiver and should be
	// identical in both.
	// Refer to section 4.1.1.6 in the datasheet for more information.
	PreambleLength uint

	// HighPower specifies whether we are to use large
	// transmitting output powers for communication.
	HighPower bool

	// Agc specifies whether we should enable
	// Automatic Gain Control on the radio.
	Agc bool

	// Crc specifies whether to append a Cyclic Redundancy Check
	// on the transmiter and whether to check the packets against
	// it on the receiver.
	Crc bool

	// LogLevel controls how 'verbosy' the instantiated device is.
	LogLevel Log_level
}

// DefaultOpts are the recommended options for the radio.
var DefaultOpts = Opts{
	BaudrateMHz:    5,
	ResetPin:       rpi.P1_22,
	FrequencyMHz:   915,
	PreambleLength: 8,
	HighPower:      true,
	Agc:            false,
	Crc:            true,
	LogLevel:       LogLevelInfo,
}

// Dev represents an RFM9x radio
type Dev struct {
	// cnx is the SPI connection with the chip itself.
	cnx spi.Conn

	// rWBuff is used as the backing information source
	// and destination on SPI transactions.
	rWBuff [4]byte

	// resetPin specifies the GPIO pin physically connected
	// to the chip's reset pin.
	resetPin gpio.PinIO

	// frequencyMHz specifies the radio's carrier frequency.
	frequencyMHz int64

	// preambleLength specifies the length of
	// the preamble on generated and received packets.
	preambleLength uint

	// highPower specifies whether we're using large
	// output powers.
	highPower bool

	// agc specifies whether Automatic Gain Control is enabled.
	agc bool

	// crc specifies whether Cyclic Redundancy Checks are enabled.
	crc bool
}

// logger is used throughout the package to
// log information to stdout.
var logger rfm9x_logger

// New initialises and returns a reference to a new RFM9x radio.
//
// Configuration options are provided through o and the SPI
// port on which to communicate with the radio is provided on p.
//
// If errors are encountered during initialisation, an empty
// reference along with an error is returned.
func New(p spi.Port, o *Opts) (*Dev, error) {
	c, err := p.Connect(physic.Frequency(o.BaudrateMHz)*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}

	logger = rfm9x_logger{level: o.LogLevel}

	logger.debug("Connection state: %v\n", c.Duplex().String())

	dev := &Dev{
		cnx:            c,
		rWBuff:         [4]byte{},
		resetPin:       o.ResetPin,
		frequencyMHz:   o.FrequencyMHz,
		preambleLength: o.PreambleLength,
		highPower:      o.HighPower,
		agc:            o.Agc,
		crc:            o.Crc,
	}

	dev.Reset()
	if v, err := dev.Version(); v != 18 || err != nil {
		logger.warn("Wrong radio version detected O_o!\n")
	}

	dev.SetMode(OpModeSleep)
	time.Sleep(10 * time.Millisecond)
	logger.debug("Current operating mode: %s", OpModeText(dev.Mode()))

	dev.SetLoRa(true)
	logger.debug("LoRa mode enabled? %v\n", dev.LoRa())

	if o.FrequencyMHz > 525 {
		dev.SetLowFreqMode(false)
		logger.debug("Low frequency mode enabled? %v\n", dev.LowFreqMode())
	}

	dev.SetFifoBaseAddrs(0x0, 0x0)
	dev.SetCarrierFrequencyMHz(o.FrequencyMHz)
	dev.SetPreambleLength(uint16(o.PreambleLength))
	dev.SetBwHz(125000)
	dev.SetCodingRate(5)
	dev.SetSpreadingFactor(7)
	dev.SetCrc(o.Crc)
	dev.SetAgc(o.Agc)
	dev.SetTxPower(13)
	dev.SetMode(OpModeStandby)

	if o.LogLevel <= LogLevelDebug {
		tx_b, rx_b := dev.FifoBaseAddrs()
		logger.debug("FiFo base addresses -> Tx = %v; Rx = %v\n", tx_b, rx_b)
		m_freq, _ := dev.CarrierFrequencyMHz()
		logger.debug("Current modulating frequency: %v MHz\n", m_freq)
		pre_l, _ := dev.PreambleLength()
		logger.debug("Current preamble length frequency: %v bytes\n", pre_l)
		bw_hz, _ := dev.BwHz()
		logger.debug("Current bandwidth: %v Hz\n", bw_hz)
		cr, _ := dev.CodingRate()
		logger.debug("Current coding rate: %v\n", cr)
		sf, _ := dev.SpreadingFactor()
		logger.debug("Current spreading factor: %v \n", sf)
		logger.debug("CRC enabled? %v\n", dev.Crc())
		logger.debug("AGC enabled? %v\n", dev.Agc())
		tx_pow, _ := dev.TxPower()
		logger.debug("Current TX power: %v dBm\n", tx_pow)
		time.Sleep(10 * time.Millisecond)
		logger.debug("Current operating mode: %s", OpModeText(dev.Mode()))
	}

	return dev, nil
}

// Reset drives the radio's reset pin to
// return it to a known state. It also
// checks whether the operation was successful
// by reading back the value of a register with
// a well known default value.
func (d *Dev) Reset() {
	logger.debug("Began resetting the radio!\n")

	d.resetPin.Out(gpio.High)
	d.resetPin.Out(gpio.Low)
	time.Sleep(100 * time.Microsecond)
	d.resetPin.Out(gpio.High)
	time.Sleep(5 * time.Millisecond)

	if mode, _ := d.read_register(RegOpMode, 3, 0); mode != 0x1 {
		logger.debug("Looks like the RESET didn't work as planned: current Op Mode is %d\n", mode)
	} else {
		logger.debug("The RESET looks good!\n")
	}
}

// Version returns the chips version number
// along with any errors raised by the
// SPI transaction.
func (d *Dev) Version() (byte, error) {
	return d.read_byte(RegVersion)
}

// Print_registers shows the contents of the main configuration registers.
// It is mainly intended for debugging and checking the correctness of the
// current configuration.
func (d *Dev) Print_registers() {
	fmt.Printf("Operation mode: %s\n", fmt.Sprint(d.read_register(RegOpMode, 3, 0)))
	fmt.Printf("Low frequency mode: %s\n", fmt.Sprint(d.read_register(RegOpMode, 1, 3)))
	fmt.Printf("Modulation type: %s\n", fmt.Sprint(d.read_register(RegOpMode, 2, 5)))
	fmt.Printf("LoRa mode: %s\n", fmt.Sprint(d.read_register(RegOpMode, 1, 7)))
	fmt.Printf("Output power: %s\n", fmt.Sprint(d.read_register(RegPaConfig, 4, 0)))
	fmt.Printf("Max power: %s\n", fmt.Sprint(d.read_register(RegPaConfig, 3, 4)))
	fmt.Printf("Pa Config: %s\n", fmt.Sprint(d.read_register(RegPaConfig, 8, 0)))
	fmt.Printf("PA select: %s\n", fmt.Sprint(d.read_register(RegOpMode, 1, 7)))
	fmt.Printf("PA DAC: %s\n", fmt.Sprint(d.read_register(RegPaDac, 3, 0)))
	fmt.Printf("DIO 0 mapping: %s\n", fmt.Sprint(d.read_register(RegDioMappingA, 2, 6)))
	fmt.Printf("Auto AGC: %s\n", fmt.Sprint(d.read_register(RegModemConfigC, 1, 2)))
	fmt.Printf("Low datarate optimise: %s\n", fmt.Sprint(d.read_register(RegModemConfigC, 1, 3)))
	fmt.Printf("LNA boost HF: %s\n", fmt.Sprint(d.read_register(RegLna, 2, 0)))
	fmt.Printf("Auto IF on: %s\n", fmt.Sprint(d.read_register(RegDetectionOptimize, 1, 7)))
	fmt.Printf("Detection optimise: %s\n", fmt.Sprint(d.read_register(RegDetectionOptimize, 3, 0)))

	fmt.Printf("Raw Freq MSB register: %s\n", fmt.Sprint(d.read_register(RegFrfMsb, 8, 0)))
	fmt.Printf("Raw Freq MID registers: %s\n", fmt.Sprint(d.read_register(RegFrfMid, 8, 0)))
	fmt.Printf("Raw Freq LSB registers: %s\n", fmt.Sprint(d.read_register(RegFrfLsb, 8, 0)))
}

// read_register returns a register slice of the given size at the given offset
// from the register identified by the provided addr.
// It resturns the read data and any errors raised by the SPI transaction.
func (d *Dev) read_register(addr reg_addr, size, offset byte) (byte, error) {
	c_reg, err := d.read_byte(addr)
	if err != nil {
		return 0xFF, err
	}
	return (c_reg & (((1 << size) - 1) << offset)) >> offset, nil
}

// write_register overwrites a register slice of the given size at the
// given offset with the given data. The register is identified by its addr.
// It returns any errors raised by the SPI interface.
func (d *Dev) write_register(addr reg_addr, size, offset, data byte) error {
	c_reg, err := d.read_byte(addr)
	if err != nil {
		return err
	}

	// Clear out the register part we are to modify
	c_reg &= (^((1 << size) - 1)) << offset

	// Force that part to the provided data
	c_reg |= (data & 0xFF) << offset

	return d.write_byte(addr, c_reg)
}

// read_byte retrieves a byte located at the provided
// addr over a SPI connection.
// It returns the read data and any errors raised by the
// SPI transaction.
func (d *Dev) read_byte(addr reg_addr) (byte, error) {
	d.rWBuff[0] = byte(addr) & 0x7F
	if err := d.cnx.Tx(d.rWBuff[:2], d.rWBuff[:2]); err != nil {
		return 0xFF, err
	}
	logger.reg_io("READ  @ Address -> %v; R/W buffer -> %v\n", addr, d.rWBuff)
	return d.rWBuff[1], nil
}

// write_byte writes the specified data at the provided addr
// over a SPI connection.
// It returns any errors raised by the SPI transaction.
func (d *Dev) write_byte(addr reg_addr, data byte) error {
	d.rWBuff[0] = (byte(addr) | 0x80) & 0xFF
	d.rWBuff[1] = data & 0xFF
	if err := d.cnx.Tx(d.rWBuff[:2], d.rWBuff[:2]); err != nil {
		return err
	}
	logger.reg_io("WRITE @ Address -> %v; R/W buffer -> %v\n", addr, d.rWBuff)
	return nil
}

// write_payload writes a stream of bytes (i.e. data) at the
// provided addr in a single SPI transaction. This permits
// keeping the SS line low instead of toggling it back and forth.
// It returns any errors raised by the SPI transaction.
func (d *Dev) write_payload(addr byte, data []byte) error {
	payload := []byte{(byte(addr) | 0x80) & 0xFF}
	payload = append(payload, data...)
	recv := make([]byte, len(payload))
	if err := d.cnx.Tx(payload, recv); err != nil {
		return err
	}
	return nil
}
