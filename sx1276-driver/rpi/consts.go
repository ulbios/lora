package rfm9x

type op_mode byte
type reg_addr byte

const (
	// Configuration register addresses.
	// See `https://go.dev/src/net/http/status.go` for an example from the Go authors.
	RegFifo                reg_addr = 0x00
	RegOpMode              reg_addr = 0x01
	RegFrfMsb              reg_addr = 0x06
	RegFrfMid              reg_addr = 0x07
	RegFrfLsb              reg_addr = 0x08
	RegPaConfig            reg_addr = 0x09
	RegPaRamp              reg_addr = 0x0A
	RegOcp                 reg_addr = 0x0B
	RegLna                 reg_addr = 0x0C
	RegFifoAddrPtr         reg_addr = 0x0D
	RegFifoTxBaseAddr      reg_addr = 0x0E
	RegFifoRxBaseAddr      reg_addr = 0x0F
	RegFifoRxCurrentAddr   reg_addr = 0x10
	RegIrqFlagsMask        reg_addr = 0x11
	RegIrqFlags            reg_addr = 0x12
	RegRxNbBytes           reg_addr = 0x13
	RegRxHeaderCntValueMsb reg_addr = 0x14
	RegRxHeaderCntValueLsb reg_addr = 0x15
	RegRxHacketCntValueMsb reg_addr = 0x16
	RegRxHacketCntValueLsb reg_addr = 0x17
	RegModemStat           reg_addr = 0x18
	RegPktSnrValue         reg_addr = 0x19
	RegPktRssiValue        reg_addr = 0x1A
	RegRssiValue           reg_addr = 0x1B
	RegHopChannel          reg_addr = 0x1C
	RegModemConfigA        reg_addr = 0x1D
	RegModemConfigB        reg_addr = 0x1E
	RegSymbTimeoutLsb      reg_addr = 0x1F
	RegPreambleMsb         reg_addr = 0x20
	RegPreambleLsb         reg_addr = 0x21
	RegPayloadLength       reg_addr = 0x22
	RegMaxPayloadLength    reg_addr = 0x23
	RegHopPeriod           reg_addr = 0x24
	RegFifoRxByteAddr      reg_addr = 0x25
	RegModemConfigC        reg_addr = 0x26
	RegDioMappingA         reg_addr = 0x40
	RegDioMappingB         reg_addr = 0x41
	RegVersion             reg_addr = 0x42
	RegTcxo                reg_addr = 0x4B
	RegPaDac               reg_addr = 0x4D
	RegFormerTemp          reg_addr = 0x5B
	RegAgcRef              reg_addr = 0x61
	RegAgcThreshA          reg_addr = 0x62
	RegAgcThreshB          reg_addr = 0x63
	RegAgcThreshC          reg_addr = 0x64
	RegDetectionOptimize   reg_addr = 0x31
	RegDetectionThreshold  reg_addr = 0x37

	// Check table 42 on the datasheet for information on
	// the mapping of operating modes.
	OpModeSleep   op_mode = 0b000
	OpModeStandby op_mode = 0b001
	OpModeFsTx    op_mode = 0b010
	OpModeTx      op_mode = 0b011
	OpModeFsRx    op_mode = 0b100
	OpModeRx      op_mode = 0b101

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
	opModeText = map[op_mode]string{
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
func OpModeText(m op_mode) string {
	return opModeText[m]
}
