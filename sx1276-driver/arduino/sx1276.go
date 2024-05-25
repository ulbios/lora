package arduino

import (
	"machine"
	"time"
)

// Opts defines configurable options for the device.
type Opts struct {
	// BaudRateMHZ specifies the baudrate at which the SPI
	// 'dialog' happens. This value is assumed to be provided
	// in MegaHertz (i.e. MHz).
	Baudrate uint32

	// LittleEndian controls whether the SPI interface's endianness
	LittleEndian bool

	// Mode controls the transmission mode (i.e. when data is shifted).
	// Check https://docs.arduino.cc/learn/communication/spi for more
	// information.
	Mode uint8

	// ResetPin specifies the physical GPIO pin the
	// chip's RESET input is connected to.
	ResetPin machine.Pin

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
	// on the transmitter and whether to check the packets against
	// it on the receiver.
	Crc bool

	BandwidthKHz    uint
	CodingRate      byte
	SpreadingFactor byte
	TxPowerDbm      uint
}

// DefaultOpts are the recommended options for the radio.
var DefaultOpts = Opts{
	Baudrate:        5 * machine.MHz,
	LittleEndian:    false,
	Mode:            machine.Mode0,
	ResetPin:        machine.D9,
	FrequencyMHz:    868,
	PreambleLength:  8,
	HighPower:       true,
	Agc:             false,
	Crc:             true,
	BandwidthKHz:    125 * machine.KHz,
	CodingRate:      8,
	SpreadingFactor: 12,
	TxPowerDbm:      13,
}

// Dev represents an RFM9x radio
type Dev struct {
	// s is the SPI instance representing the interface.
	s machine.SPI

	// rWBuff is used as the backing information source
	// and destination on SPI transactions.
	rWBuff []byte

	slaveSelectPin machine.Pin

	// resetPin specifies the GPIO pin physically connected
	// to the chip's reset pin.
	resetPin machine.Pin

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

// New initialises and returns a reference to a new RFM9x radio.
//
// Configuration options are provided through o and the SPI
// port on which to communicate with the radio is provided on p.
//
// If errors are encountered during initialisation, an empty
// reference along with an error is returned.
func New(o *Opts) (*Dev, error) {
	s := machine.SPI0
	if err := s.Configure(
		machine.SPIConfig{Frequency: o.Baudrate, LSBFirst: o.LittleEndian, Mode: o.Mode}); err != nil {
		println("error configuring the SPI port:", err)
	}

	o.ResetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	dev := &Dev{
		s:              s,
		rWBuff:         make([]byte, 4),
		resetPin:       o.ResetPin,
		slaveSelectPin: machine.D10,
		frequencyMHz:   o.FrequencyMHz,
		preambleLength: o.PreambleLength,
		highPower:      o.HighPower,
		agc:            o.Agc,
		crc:            o.Crc,
	}

	dev.Reset()
	if v, err := dev.Version(); v != 18 || err != nil {
		println("Wrong radio version detected!", v, err)
	}

	dev.SetMode(OpModeSleep)
	time.Sleep(10 * time.Millisecond)
	println("Current operating mode: ", OpModeText(dev.Mode()))

	dev.SetLoRa(true)
	println("LoRa mode enabled? ", dev.LoRa())

	if o.FrequencyMHz > 525 {
		dev.SetLowFreqMode(false)
		println("Low frequency mode enabled? ", dev.LowFreqMode())
	}

	dev.SetFifoBaseAddrs(0x0, 0x0)
	dev.SetCarrierFrequencyMHz(o.FrequencyMHz)
	dev.SetPreambleLength(uint16(o.PreambleLength))
	dev.SetBwHz(o.BandwidthKHz)
	dev.SetCodingRate(o.CodingRate)
	dev.SetSpreadingFactor(o.SpreadingFactor)
	dev.SetCrc(o.Crc)
	dev.SetAgc(o.Agc)
	dev.SetTxPower(o.TxPowerDbm)
	dev.SetMode(OpModeStandby)

	txB, rxB := dev.FifoBaseAddrs()
	println("FiFo base addresses -> Tx = ", txB, " Rx = ", rxB)

	mFreq, _ := dev.CarrierFrequencyMHz()
	println("Current modulating frequency: ", mFreq, " MHz")

	preL, _ := dev.PreambleLength()
	println("Current preamble length frequency: ", preL, " bytes")

	bwHz, _ := dev.BwHz()
	println("Current bandwidth: ", bwHz, " Hz")

	cR, _ := dev.CodingRate()
	println("Current coding rate: ", cR)

	sF, _ := dev.SpreadingFactor()
	println("Current spreading factor:", sF)
	println("CRC enabled? ", dev.Crc())
	println("AGC enabled? ", dev.Agc())

	txPow, _ := dev.TxPower()
	println("Current TX power: ", txPow, " dBm")
	time.Sleep(10 * time.Millisecond)
	println("Current operating mode: ", OpModeText(dev.Mode()))

	return dev, nil
}

// Reset drives the radio's reset pin to
// return it to a known state. It also
// checks whether the operation was successful
// by reading back the value of a register with
// a well known default value.
func (d *Dev) Reset() {
	// d.resetPin.High()
	d.resetPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(100 * time.Millisecond)

	if mode, _ := d.readRegister(RegOpMode, 3, 0); mode != 0x1 {
		println("Looks like the RESET didn't work as planned: current Op Mode is ", mode)
	} else {
		println("The RESET looks good!")
	}
}

// Version returns the chips version number
// along with any errors raised by the
// SPI transaction.
func (d *Dev) Version() (byte, error) {
	return d.readByte(RegVersion)
}
