# Modbus Data Gateway
This directory contains the implementation of a LoRa-based radio gateway. The distances
we had to cover at Guadalajara's EDAR proved to be much longer than initially forseen.
In order to reliably transmit data from the farthest sensor to the control hub we
designed a gateway sitting in the middle. This gateway would relay information from the
farthest sensor whilst also sending its own sensor's data. This project is intended to use a
RFM9x LoRa radio.

This gateway radiated data so that it would be received at the EDAR's control hub, where the
implementation residing on `../mb-server` made the information available to other equipment.
Thus, the overall system could be regarded as:

    + ---------- +       + ---------- +       + --------- +
    | mb-emitter | ----> | mb-gateway | ----> | mb-server |
    + ---------- +       + ---------- +       + --------- +

The project is intended to be ran as a headless daemon through the use of the provided SystemD
unit define on file `mb-gateway.service`.

As usual, compilation can be achieved with:

    $ GOOS=linux GOARCH=arm go build

Please note this project is not currently deployed anywhere: it has been superseded by the
implementation residing on `../mb-master`.
