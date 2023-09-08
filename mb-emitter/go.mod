module github.com/ulbios/lora/mb-emitter

go 1.17

replace ulbios/rfm9x-driver => /Users/collado/Repos/ulbios/ulbios_lora/driver/rfm9x_driver

require (
	github.com/go-co-op/gocron v1.13.0
	github.com/grid-x/modbus v0.0.0-20220419073012-0daecbb3900f
	github.com/spf13/cobra v1.4.0
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/host/v3 v3.7.2
	ulbios/rfm9x-driver v0.0.0-00010101000000-000000000000
)

require (
	github.com/grid-x/serial v0.0.0-20191104121038-e24bc9bf6f08 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
