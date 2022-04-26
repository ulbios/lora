module mb-emitter

go 1.17

replace ulbios/rfm9x-driver => /Users/collado/Repos/ulbios/ulbios_lora/driver/rfm9x_driver

require (
	github.com/grid-x/modbus v0.0.0-20220419073012-0daecbb3900f
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/host/v3 v3.7.2
	ulbios/rfm9x-driver v0.0.0-00010101000000-000000000000
)

require github.com/grid-x/serial v0.0.0-20191104121038-e24bc9bf6f08 // indirect
