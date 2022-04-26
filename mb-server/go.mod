module mb-server

go 1.17

replace ulbios/rfm9x-driver => /Users/collado/Repos/ulbios/ulbios_lora/driver/rfm9x_driver

require (
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0
	github.com/wilkingj/GoModbusServer v0.0.0-20181106112653-9397ee43cc9a
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/host/v3 v3.7.2
	ulbios/rfm9x-driver v0.0.0-00010101000000-000000000000
)

require github.com/TheCount/modbus v0.0.0-20180823092113-392130db12d5 // indirect
