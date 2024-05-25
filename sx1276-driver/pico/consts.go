package pico

type opMode byte
type regAddr byte

const (
	// Configuration register addresses.
	// See `https://go.dev/src/net/http/status.go` for an example from the Go authors.
	RegFifo                regAddr = 0x00
	RegOpMode              regAddr = 0x01
	RegFrfMsb              regAddr = 0x06
	RegFrfMid              regAddr = 0x07
	RegFrfLsb              regAddr = 0x08
	RegPaConfig            regAddr = 0x09
	RegPaRamp              regAddr = 0x0A
	RegOcp                 regAddr = 0x0B
	RegLna                 regAddr = 0x0C
	RegFifoAddrPtr         regAddr = 0x0D
	RegFifoTxBaseAddr      regAddr = 0x0E
	RegFifoRxBaseAddr      regAddr = 0x0F
	RegFifoRxCurrentAddr   regAddr = 0x10
	RegIrqFlagsMask        regAddr = 0x11
	RegIrqFlags            regAddr = 0x12
	RegRxNbBytes           regAddr = 0x13
	RegRxHeaderCntValueMsb regAddr = 0x14
	RegRxHeaderCntValueLsb regAddr = 0x15
	RegRxHacketCntValueMsb regAddr = 0x16
	RegRxHacketCntValueLsb regAddr = 0x17
	RegModemStat           regAddr = 0x18
	RegPktSnrValue         regAddr = 0x19
	RegPktRssiValue        regAddr = 0x1A
	RegRssiValue           regAddr = 0x1B
	RegHopChannel          regAddr = 0x1C
	RegModemConfigA        regAddr = 0x1D
	RegModemConfigB        regAddr = 0x1E
	RegSymbTimeoutLsb      regAddr = 0x1F
	RegPreambleMsb         regAddr = 0x20
	RegPreambleLsb         regAddr = 0x21
	RegPayloadLength       regAddr = 0x22
	RegMaxPayloadLength    regAddr = 0x23
	RegHopPeriod           regAddr = 0x24
	RegFifoRxByteAddr      regAddr = 0x25
	RegModemConfigC        regAddr = 0x26
	RegDioMappingA         regAddr = 0x40
	RegDioMappingB         regAddr = 0x41
	RegVersion             regAddr = 0x42
	RegTcxo                regAddr = 0x4B
	RegPaDac               regAddr = 0x4D
	RegFormerTemp          regAddr = 0x5B
	RegAgcRef              regAddr = 0x61
	RegAgcThreshA          regAddr = 0x62
	RegAgcThreshB          regAddr = 0x63
	RegAgcThreshC          regAddr = 0x64
	RegDetectionOptimize   regAddr = 0x31
	RegDetectionThreshold  regAddr = 0x37

	// Check table 42 on the datasheet for information on
	// the mapping of operating modes.
	OpModeSleep   opMode = 0b000
	OpModeStandby opMode = 0b001
	OpModeFsTx    opMode = 0b010
	OpModeTx      opMode = 0b011
	OpModeFsRx    opMode = 0b100
	OpModeRx      opMode = 0b101

	// Oscilator frequency. Check section 3 on the datasheet
	OscFreqHz int64 = 32000000

	// Conversion factor for carrier frequency configuration.
	// Check section 4.1.4 on the datasheet for details.
	FStepHz int64 = OscFreqHz / 524288 // 524288 = 2^19

	// Values for enabling or disabling the Power Amplifier's features.
	PaDacEnable  byte = 0x7
	PaDacDisable byte = 0x4
)

var (
	// BWID2Hz allows us to translate bandwidth IDs to the appropriate frequencies in Hertzs (i.e. Hz).
	BWID2Hz [9]uint = [9]uint{7800, 10400, 15600, 20800, 31250, 41700, 62500, 125000, 250000}

	// opModeText allows us to translate numeric operation modes into
	// user-friendly strings suitable for textual output.
	opModeText = map[opMode]string{
		OpModeSleep:   "Sleep",
		OpModeStandby: "Standby",
		OpModeFsTx:    "FsTx",
		OpModeTx:      "Tx",
		OpModeFsRx:    "FsRx",
		OpModeRx:      "Rx",
	}

	// boolToByte maps boolean values to a byte so that we can make writes
	// more succinct when configuring boolean properties of the chip.
	boolToByte = map[bool]byte{
		false: 0x0,
		true:  0x1,
	}
)

// OpModeText wraps the opModeText map so that it can be safely
// leveraged from the 'outside world'.
func OpModeText(m opMode) string {
	return opModeText[m]
}
