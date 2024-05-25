/*
Package arduino implements a driver for Adafruit's RFM9x radios.
These are based on Semtech's SX1276/77/78/79 radio transceiver.

The target device is an Arduino Uno board whose firmware is compiled
with the TinyGo infrastructure.

Useful resources:

	Datasheet: https://cdn-shop.adafruit.com/product-files/3179/sx1276_77_78_79.pdf
	Errata: https://www.digchip.com/datasheets/download_datasheet.php?id=8756311&part-number=SX1276RF1KAS
	Reference Python implementation: https://github.com/adafruit/Adafruit_CircuitPython_RFM9x/blob/main/adafruit_rfm9x.py
	Reference C++ implementation: https://github.com/mirtcho/LoRa/blob/master/src/LoRa.cpp
*/
package pico
