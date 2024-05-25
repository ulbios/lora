package pico

import (
	"errors"
	"time"
)

// TxDone returns a boolean indicating whether the Tx IRQ
// flag is set or not. If the underlying SPI transcation
// throws an error we'll default to assuming the Tx ins't
// finished, thus returning false.
func (d *Dev) TxDone() bool {
	tx_flag, err := d.readRegister(RegIrqFlags, 1, 3)
	if err != nil {
		return false
	}
	return tx_flag == 0x1
}

// RxDone returns a boolean indicating whether the Rx IRQ
// flag is set or not. If the underlying SPI transcation
// throws an error we'll default to assuming the Rx ins't
// finished, thus returning false.
func (d *Dev) RxDone() bool {
	rx_flag, err := d.readRegister(RegIrqFlags, 1, 6)
	if err != nil {
		return false
	}
	return rx_flag == 0x1
}

// Send transmits the data provided on data. The radio will be
// transitioned to Tx mode and then returned back to Standby
// once the transmission is finished. As of now, the call blocks
// until the transmission finishes, possibly causing a deadlock.
// It returns any errors triggered by the underlying SPI
// transactions.
func (d *Dev) Send(data []byte) error {
	d.SetMode(OpModeStandby)
	println("# COMMS # Current operating mode: ", OpModeText(d.Mode()))

	d.writeRegister(RegFifoAddrPtr, 8, 0, 0x0)
	rh_header := []byte{0xFF, 0xFF, 0x0, 0x0}
	payload := append(rh_header, data...)
	d.writePayload(byte(RegFifo), payload)

	println("# COMMS # Wrote ", payload, " to the FiFo with length ", len(payload))
	d.writeRegister(RegPayloadLength, 8, 0, byte(len(payload)))

	d.SetMode(OpModeTx)
	d.writeRegister(RegDioMappingA, 2, 6, 0x1)
	// time.Sleep(500 * time.Millisecond)
	// logger.debug("# COMMS # Current operating mode: %s\n", OpModeText(d.Get_mode()))
	println("# COMMS # Current operating mode: ", OpModeText(d.Mode()))

	for !d.TxDone() {
		println("# COMMS # Sending hasn't been ACKd yet...")
		time.Sleep(1 * time.Second)
	}
	println("# COMMS # Looks like they've ACKd us!")

	d.SetMode(OpModeStandby)

	// Clear IRQs
	d.writeRegister(RegIrqFlags, 8, 0, 0xFF)

	return nil
}

func (d *Dev) Receive(wait, timeout time.Duration) ([]byte, error) {
	println("# COMMS # Beginning to listen for a packet")
	d.SetMode(OpModeRx)

	var timeWaited time.Duration = 0
	for !d.RxDone() {
		time.Sleep(wait)
		println("# COMMS # Waiting for another ", wait)
		timeWaited += wait
		if timeout != 0 && timeWaited >= timeout {
			d.writeRegister(RegIrqFlags, 8, 0, 0xFF)
			d.SetMode(OpModeStandby)
			return nil, errors.New("timeout on reception")
		}
	}

	pktLen, _ := d.readRegister(RegRxNbBytes, 8, 0)
	println("# COMMS # Received a ", pktLen, "-bit bytes long packet!")

	if pktLen == 0 {
		d.writeRegister(RegIrqFlags, 8, 0, 0xFF)
		d.SetMode(OpModeStandby)
		return nil, errors.New("received an empty packet")
	}

	pkt_addr, _ := d.readRegister(RegFifoRxCurrentAddr, 8, 0)
	d.writeRegister(RegFifoAddrPtr, 8, 0, pkt_addr)

	pkt := make([]byte, pktLen)
	for i := 0; i < int(pktLen); i++ {
		b, _ := d.readRegister(RegFifo, 8, 0)
		pkt[i] = b
	}

	println("# COMMS # Received ", pkt, " from the FiFo with length ", len(pkt))

	// Clear IRQs
	d.writeRegister(RegIrqFlags, 8, 0, 0xFF)

	d.SetMode(OpModeStandby)

	return pkt, nil
}
