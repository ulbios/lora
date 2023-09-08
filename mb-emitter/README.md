# Modbus Data Emitter over LoRa
This directory contains the implementation of a Modbus client capable of radiating
retrieved information over LoRa through a RFM9x LoRa radio.

This project was deployed on the field on Guadalajara's EDAR so that data from
depth sensors could be transmitted to a control hub more than 500 meters away.

The data emitted by this implementation was to be picked up at the EDAR's
control hub by a running instance of the implementation living on `../mb-server`.
After we ran into some coverage issues, these data points were to be received
at an intermediate point by a running instance of the `../mb-gateway` implementation.

This implementation is intended to be executed as a hedaless daemon by leveraging
the SystemD unit files provided in this same directory. File `mb-emitter.service`
should be used on Raspberry Pi's, whilst file `mb-emitter-opi.service` should be
leveraged when running on Orange Pi's (i.e. a Raspberry Pi clone).

As usual, compilation can be achieved with:

    $ GOOS=linux GOARCH=arm go build

Please note this project is not currently deployed anywhere: it has been superseded
by the implementation residing on `../mb-master`.
