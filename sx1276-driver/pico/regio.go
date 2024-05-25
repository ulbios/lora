package pico

// read_register returns a register slice of the given size at the given offset
// from the register identified by the provided addr.
// It resturns the read data and any errors raised by the SPI transaction.
func (d *Dev) readRegister(addr regAddr, size, offset byte) (byte, error) {
	c_reg, err := d.readByte(addr)
	println("read byte", c_reg)
	if err != nil {
		return 0xFF, err
	}
	return (c_reg & (((1 << size) - 1) << offset)) >> offset, nil
}

// write_register overwrites a register slice of the given size at the
// given offset with the given data. The register is identified by its addr.
// It returns any errors raised by the SPI interface.
func (d *Dev) writeRegister(addr regAddr, size, offset, data byte) error {
	c_reg, err := d.readByte(addr)
	if err != nil {
		return err
	}

	// Clear out the register part we are to modify
	c_reg &= (^((1 << size) - 1)) << offset

	// Force that part to the provided data
	c_reg |= (data & 0xFF) << offset

	return d.writeByte(addr, c_reg)
}

// read_byte retrieves a byte located at the provided
// addr over a SPI connection.
// It returns the read data and any errors raised by the
// SPI transaction.
func (d *Dev) readByte(addr regAddr) (byte, error) {
	d.rWBuff[0] = byte(addr) & 0x7F
	d.slaveSelectPin.Low()
	if err := d.s.Tx(d.rWBuff[:2], d.rWBuff[:2]); err != nil {
		return 0xFF, err
	}
	d.slaveSelectPin.High()
	// println("READ @ Address -> ", addr, "R/W buffer -> ", d.rWBuff)
	return d.rWBuff[1], nil
}

// write_byte writes the specified data at the provided addr
// over a SPI connection.
// It returns any errors raised by the SPI transaction.
func (d *Dev) writeByte(addr regAddr, data byte) error {
	d.rWBuff[0] = (byte(addr) | 0x80) & 0xFF
	d.rWBuff[1] = data & 0xFF
	d.slaveSelectPin.Low()
	if err := d.s.Tx(d.rWBuff[:2], nil); err != nil {
		return err
	}
	d.slaveSelectPin.High()
	// println("WRITE @ Address -> ", addr, "R/W buffer -> ", d.rWBuff)
	return nil
}

// write_payload writes a stream of bytes (i.e. data) at the
// provided addr in a single SPI transaction. This permits
// keeping the SS line low instead of toggling it back and forth.
// It returns any errors raised by the SPI transaction.
func (d *Dev) writePayload(addr byte, data []byte) error {
	payload := []byte{(byte(addr) | 0x80) & 0xFF}
	payload = append(payload, data...)
	recv := make([]byte, len(payload))
	d.slaveSelectPin.Low()
	if err := d.s.Tx(payload, recv); err != nil {
		return err
	}
	d.slaveSelectPin.High()
	return nil
}
