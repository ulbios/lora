package rfm9x

import (
	"fmt"
	"time"
)

// TxDone returns a boolean indicating whether the Tx IRQ
// flag is set or not. If the underlying SPI transcation
// throws an error we'll default to assuming the Tx ins't
// finished, thus returning false.
func (d *Dev) TxDone() bool {
	tx_flag, err := d.read_register(RegIrqFlags, 1, 3)
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
	rx_flag, err := d.read_register(RegIrqFlags, 1, 6)
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
	logger.debug("# COMMS # Current operating mode: %s\n", OpModeText(d.Mode()))

	d.write_register(RegFifoAddrPtr, 8, 0, 0x0)
	rh_header := []byte{0xFF, 0xFF, 0x0, 0x0}
	payload := append(rh_header, data...)
	d.write_payload(byte(RegFifo), payload)

	logger.debug("# COMMS # Wrote %v to the FiFo [length = %v]\n", payload, byte(len(payload)))
	d.write_register(RegPayloadLength, 8, 0, byte(len(payload)))

	d.SetMode(OpModeTx)
	d.write_register(RegDioMappingA, 2, 6, 0x1)
	// time.Sleep(500 * time.Millisecond)
	// logger.debug("# COMMS # Current operating mode: %s\n", OpModeText(d.Get_mode()))
	logger.debug("# COMMS # Current operating mode: %s\n", fmt.Sprint(d.read_register(RegOpMode, 8, 0)))

	for !d.TxDone() {
		logger.debug("# COMMS # Sending hasn't been ACKd yet...\n")
		time.Sleep(1 * time.Second)
	}
	logger.debug("# COMMS # Looks like they've ACKd us!\n")

	d.SetMode(OpModeStandby)

	// Clear IRQs
	d.write_register(RegIrqFlags, 8, 0, 0xFF)

	return nil
}

func (d *Dev) Receive(wait, timeout time.Duration) ([]byte, error) {
	logger.debug("# COMMS # Beginning to listen for a packet\n")
	d.SetMode(OpModeRx)

	var timeWaited time.Duration = 0
	for !d.RxDone() {
		time.Sleep(wait)
		logger.debug("# COMMS # Waiting for another %v...\n", wait)
		timeWaited += wait
		if timeout != 0 && timeWaited >= timeout {
			d.write_register(RegIrqFlags, 8, 0, 0xFF)
			d.SetMode(OpModeStandby)
			return nil, fmt.Errorf("timeout on reception")
		}
	}

	pkt_len, _ := d.read_register(RegRxNbBytes, 8, 0)
	logger.debug("# COMMS # Received a %d-bit bytes long packet!", pkt_len)

	if pkt_len == 0 {
		d.write_register(RegIrqFlags, 8, 0, 0xFF)
		d.SetMode(OpModeStandby)
		return nil, fmt.Errorf("received an empty packet")
	}

	pkt_addr, _ := d.read_register(RegFifoRxCurrentAddr, 8, 0)
	d.write_register(RegFifoAddrPtr, 8, 0, pkt_addr)

	pkt := make([]byte, pkt_len)
	for i := 0; i < int(pkt_len); i++ {
		b, _ := d.read_register(RegFifo, 8, 0)
		pkt[i] = b
	}

	logger.debug("# COMMS # Received %v from the FiFo [length = %v]\n", pkt, len(pkt))

	// Clear IRQs
	d.write_register(RegIrqFlags, 8, 0, 0xFF)

	d.SetMode(OpModeStandby)

	return pkt, nil
}
