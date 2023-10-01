package arduino

import (
	"errors"
)

// Mode returns the current operation mode of the radio.
// External users can leverage the OpModeText function
// to convert it to a readable string.
func (d *Dev) Mode() opMode {
	m, err := d.readRegister(RegOpMode, 3, 0)
	if err != nil {
		return 0x7
	}
	return opMode(m)
}

// SetMode transitions the chip to the provided operation mode.
// It resturns any errors raised by the underlying SPI transaction.
func (d *Dev) SetMode(mode opMode) error {
	return d.writeRegister(RegOpMode, 3, 0, byte(mode))
}

// LowFreqMode returns a boolean indicating whether the radio is
// currently configured to use the low or high frequency registers.
// If the underlying SPI transaction raises an error, false will
// always be returned.
func (d *Dev) LowFreqMode() bool {
	lf_mode, err := d.readRegister(RegOpMode, 1, 3)
	if err != nil {
		return false
	}
	return lf_mode == 0x1
}

// SetLowFreqMode configures the chip to use either low of high
// frequency registers according to the value of enable.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetLowFreqMode(enable bool) error {
	return d.writeRegister(RegOpMode, 1, 3, boolToByte[enable])
}

// LoRa returns a boolean indicating whether the radio is
// configured to use LoRa for transmitting information.
// If the underlying SPI transaction raises an error, false will
// always be returned.
func (d *Dev) LoRa() bool {
	opMode, err := d.readRegister(RegOpMode, 1, 7)
	if err != nil {
		return false
	}
	return opMode == 0x1
}

// SetLora configures the chip to use LoRa for sending data
// depending on the value of enable.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetLoRa(enable bool) error {
	return d.writeRegister(RegOpMode, 1, 7, boolToByte[enable])
}

// CarrierFrequencyMHz returns an integer indicating the current
// carrier frequency in MegaHertz (i.e. MHz).
// It also returns any errors raised by the underlying SPI transaction.
func (d *Dev) CarrierFrequencyMHz() (int, error) {
	msb, err := d.readRegister(RegFrfMsb, 8, 0)
	if err != nil {
		return -1, nil
	}
	mid, err := d.readRegister(RegFrfMid, 8, 0)
	if err != nil {
		return -1, err
	}
	lsb, err := d.readRegister(RegFrfLsb, 8, 0)
	if err != nil {
		return -1, err
	}

	carrier_f := int64((uint(msb)<<16)|(uint(mid)<<8)|uint(lsb)) & 0xFFFFFF

	// Refer to section 4.1.4 for a justification of the following expression.
	// Note the trailing division by 10^6 accounts for a Hz -> MHz conversion.
	return int((carrier_f*OscFreqHz)>>19) / 1000000, nil
}

// SetCarrierFrequencyMHz configures the carrier frequency provided on
// carrier_f as the one used by the radio. It is assumed to be in MHz.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetCarrierFrequencyMHz(carrier_f int64) error {
	if carrier_f < 240 || carrier_f > 920 {
		return errors.New("frequency must belong to the [240, 920] MHz interval")
	}

	// Refer to section 4.1.4 for a justification of the following expression.
	// Note the initial multiplication by 10^6 accounts for a MHz -> Hz conversion.
	var frf int64 = (((carrier_f * 1000000) << 19) / OscFreqHz) & 0xFFFFFF

	if err := d.writeRegister(RegFrfMsb, 8, 0, byte(frf>>16)); err != nil {
		return err
	}

	if err := d.writeRegister(RegFrfMid, 8, 0, byte(frf>>8)); err != nil {
		return err
	}

	if err := d.writeRegister(RegFrfLsb, 8, 0, byte(frf)); err != nil {
		return err
	}

	return nil
}

// PreambleLength returns the current preamble length.
// It also returns any errors raised by the
// underlying SPI transaction.
func (d *Dev) PreambleLength() (uint16, error) {
	msb, err := d.readRegister(RegPreambleMsb, 8, 0)
	if err != nil {
		return 0, err
	}

	lsb, err := d.readRegister(RegPreambleLsb, 8, 0)
	if err != nil {
		return 0, err
	}

	return uint16(msb)<<8 | uint16(lsb), nil
}

// SetPreambleLength configures the preamble length provided on
// ln as the one used by the radio.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetPreambleLength(ln uint16) error {
	if err := d.writeRegister(RegPreambleMsb, 8, 0, byte(ln>>8)); err != nil {
		return err
	}

	return d.writeRegister(RegPreambleLsb, 8, 0, byte(ln))
}

// CodingRate returns the current coding rate.
// It also returns any errors raised by the
// underlying SPI transaction.
func (d *Dev) CodingRate() (byte, error) {
	cr_id, err := d.readRegister(RegModemConfigA, 3, 1)
	if err != nil {
		return 0, err
	}

	return cr_id + 4, nil
}

// SetCodingRate configures the coding rate provided on
// cr as the one used by the radio.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetCodingRate(cr byte) error {
	if cr < 5 || cr > 8 {
		return errors.New("incorrect coding rate id")
	}
	return d.writeRegister(RegModemConfigA, 3, 1, cr-4)
}

// SpreadingFactor returns the current spreading factor.
// It also returns any errors raised by the
// underlying SPI transaction.
func (d *Dev) SpreadingFactor() (byte, error) {
	sf, err := d.readRegister(RegModemConfigB, 4, 4)
	if err != nil {
		return 0, err
	}
	return sf, nil
}

// SetSpreadingFactor configures the spreading factor provided on
// sf as the one used by the radio.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetSpreadingFactor(sf byte) error {
	if sf < 6 || sf > 12 {
		return errors.New("incorrect spreading factor")
	}
	if sf == 6 {
		if err := d.writeRegister(RegDetectionOptimize, 3, 0, 0x5); err != nil {
			return err
		}
		if err := d.writeRegister(RegDetectionThreshold, 8, 0, 0x0C); err != nil {
			return err
		}
	} else {
		if err := d.writeRegister(RegDetectionOptimize, 3, 0, 0x3); err != nil {
			return err
		}
		if err := d.writeRegister(RegDetectionThreshold, 8, 0, 0x0A); err != nil {
			return err
		}
	}
	return d.writeRegister(RegModemConfigB, 4, 4, sf)
}

// Crc returns a boolean indicating whether
// CRC is enabled for incoming and outgoing packets.
// If the underlying SPI transaction raises an error, false will
// always be returned.
func (d *Dev) Crc() bool {
	crc, err := d.readRegister(RegModemConfigB, 1, 2)
	if err != nil {
		return false
	}
	return crc == 0x1
}

// SetCrc configures the chip to include a CRC on incoming
// and outgoing packets depending on the value of enable.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetCrc(enable bool) error {
	return d.writeRegister(RegModemConfigB, 1, 2, boolToByte[enable])
}

// Agc returns a boolean indicating whether
// Automatic Gain Control is enabled.
// If the underlying SPI transaction raises an error, false will
// always be returned.
func (d *Dev) Agc() bool {
	agc, err := d.readRegister(RegModemConfigC, 1, 2)
	if err != nil {
		return false
	}

	return agc == 0x1
}

// SetAgc configures the chip to use Automatic Gain Control
// depending on the value of enable.
// It returns any errors raised by the underlying SPI transaction.
func (d *Dev) SetAgc(enable bool) error {
	return d.writeRegister(RegModemConfigC, 1, 2, boolToByte[enable])
}

// TxPower returns the current transmission power in dBm.
// It also returns any errors raised by the underlying SPI transaction.
func (d *Dev) TxPower() (byte, error) {
	o_pow, err := d.readRegister(RegPaConfig, 4, 0)
	if err != nil {
		return 0, err
	}

	if d.highPower {
		return o_pow + 5, nil
	}
	return o_pow - 1, nil
}

// SetTxPower configures the transmission power provided through pow.
// It returns any errors raised by the underlying SPI transaction as
// well as those triggered by a malformed input parameter.
func (d *Dev) SetTxPower(pow uint) error {
	if d.highPower {
		if pow < 5 || pow > 23 {
			return errors.New("incorrect tx power")
		}

		if pow > 20 {
			d.writeRegister(RegPaDac, 3, 0, PaDacEnable)
			pow -= 3
		} else {
			d.writeRegister(RegPaDac, 3, 0, PaDacDisable)
		}
		d.writeRegister(RegPaConfig, 1, 7, 0x1)
		d.writeRegister(RegPaConfig, 3, 4, 0x04)
		d.writeRegister(RegPaConfig, 4, 0, byte((pow-5)&0xF))
	} else {
		if pow > 14 {
			return errors.New("incorrect tx power (should be between 5 and 23)")
		}
		d.writeRegister(RegPaConfig, 1, 7, 0x0)
		d.writeRegister(RegPaConfig, 3, 4, 0x7)
		d.writeRegister(RegPaConfig, 4, 0, byte((pow+1)&0xF))
	}
	return nil
}

// BwHz returns the current transmission bandwidth in Hz.
// It also returns any errors raised by the underlying SPI transaction.
func (d *Dev) BwHz() (uint, error) {
	bw_id, err := d.readRegister(RegModemConfigA, 4, 4)
	if err != nil {
		return 0, err
	}
	if int(bw_id) > len(BWID2Hz) {
		return 500000, nil
	}
	return BWID2Hz[bw_id], nil
}

// SetBwHz configures the transmission bandwidth based on bw [Hz].
// It returns any errors raised by the underlying SPI transaction.
// Values exceeding the maximum bandwidth will be truncated to the
// largest one available (i.e. 500 kHz).
func (d *Dev) SetBwHz(bw uint) error {
	/*
	 * Check the datasheet at:
	 * https://www.digchip.com/datasheets/download_datasheet.php?id=8756311&part-number=SX1276RF1KAS
	 * for information on the 'magic numbers' scattered throughout the function.
	 */

	var (
		c_bw  uint
		bw_id int
	)

	if bw > BWID2Hz[len(BWID2Hz)-1] {
		c_bw, bw_id = 500000, 9
	} else {
		for bw_id, c_bw = range BWID2Hz {
			if bw <= c_bw {
				break
			}
		}
	}

	if err := d.writeRegister(RegModemConfigA, 4, 4, byte(bw_id)); err != nil {
		return err
	}

	if c_bw >= 500000 {
		if err := d.writeRegister(RegDetectionOptimize, 1, 7, 0x1); err != nil {
			return err
		}

		if err := d.writeRegister(0x36, 8, 0, 0x02); err != nil {
			return err
		}

		l_freq_mode, err := d.readRegister(RegOpMode, 1, 3)
		if err != nil {
			return err
		}

		if l_freq_mode == 0x1 {
			if err := d.writeRegister(0x3A, 8, 0, 0x7F); err != nil {
				return err
			}
		} else {
			if err := d.writeRegister(0x3A, 8, 0, 0x64); err != nil {
				return err
			}
		}
	} else {
		if err := d.writeRegister(RegDetectionOptimize, 1, 7, 0x0); err != nil {
			return err
		}

		if err := d.writeRegister(0x36, 8, 0, 0x03); err != nil {
			return err
		}
		if c_bw == 7800 {
			if err := d.writeRegister(0x2F, 8, 0, 0x48); err != nil {
				return err
			}
		} else if c_bw >= 62500 {
			if err := d.writeRegister(0x2F, 8, 0, 0x40); err != nil {
				return err
			}
		} else {
			if err := d.writeRegister(0x2F, 8, 0, 0x44); err != nil {
				return err
			}
		}
		if err := d.writeRegister(0x30, 8, 0, 0x0); err != nil {
			return err
		}
	}

	return nil
}

// FifoBaseAddrs returns the value of the pointers indicating
// where the transmission and reception hardware FIFOs start,
// respectively. If the underlying SPI connection throws an
// error, 0xFF will be returned for the FIFO whose address
// couldn't be retrieved and a message will be logged to
// STDOUT with a warning severity.
func (d *Dev) FifoBaseAddrs() (byte, byte) {
	tx, err := d.readRegister(RegFifoTxBaseAddr, 8, 0)
	if err != nil {
		println("Error reading RegFifoTxBaseAddr!")
	}
	rx, err := d.readRegister(RegFifoRxBaseAddr, 8, 0)
	if err != nil {
		println("Error reading RegFifoRxBaseAddr!")
	}
	return tx, rx
}

// SetFifoBaseAddrs configures the hardware FIFOs to begin at the
// provided addresses. It returns any errors triggered by the
// underlying SPI transaction.
func (d *Dev) SetFifoBaseAddrs(tx, rx byte) error {
	// The chip has a single 256-bit long FIFO. We can take full advantage
	// of it by setting both the Tx and Rx addresses to 0, but we'll just
	// be multiplexing it back and forth. We could also, if needed, allocate
	// half the FIFO for Tx and the rest for Rx so as to avoid cleaning it
	// up when swapping between transceiver modes.
	if err := d.writeRegister(RegFifoTxBaseAddr, 8, 0, tx); err != nil {
		return err
	}
	if err := d.writeRegister(RegFifoRxBaseAddr, 8, 0, rx); err != nil {
		return err
	}
	return nil
}
