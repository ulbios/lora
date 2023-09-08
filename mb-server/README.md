# ModBus/{TCP, RTU} Slave over LoRa
This directory contains the implementation of a Modbus slave over both RTU (i.e. a RS-485 link) and TCP/IP. This
slave will receive data over LoRa through a RFM9x radio and then make it available at the specified addresses
based on the input options.

This implementation ran on tbe control hub at Guadalajara's EDAR where it received data from two remote
sensors leveraging the implementations found on `../mb-emitter` and `../mb-gateway`. The overall system
can be regarded as:

    + ---------- +       + ---------- +       + --------- +
    | mb-emitter | ----> | mb-gateway | ----> | mb-server |
    + ---------- +       + ---------- +       + --------- +

This implementation is to be in a headless fashion through the SystemD unit defined on file `mb-server.service`.

In order to debug the implementation, we made it possible to insert data into the Modbus slave's registers
not only through a LoRa radio, but also through UDP. That way we could verify data would be reachable by
the industrial PLCs needing to receive the data. The reason behind leaning towards UDP rests on how it is
much simpler to manage than TCP given its non-connection-oriented paradigm. What's more, thanks to Go's
builtin concurrency we could make it all work cooperatively without too much hassle.

In order to leverage our UDP data path we need to somehow send information. We can thankfully do so with `nc(1)`.
We just need to tell it to use UDP instead of TCP (which is the default) and we should be good to go:

    # Run the server on another terminal: we'll bind the UDP socket to 127.0.0.1:1503
    collado@hoth:0:~$  ./bin/mb-server -serial-device none -udp-bind-address 127.0.0.1 -udp-bind-port 1503

    # And then fire up netcat: you can type stuff on the session that opens up.
    collado@hoth:0:~$ nc -u 127.0.0.1 1503
    {"msg": "A random message xD", "data": 5}

With that you should be able to then query the server to get back the information it's updated. Nevertheless,
the server's log messages should provide very verbosy information making the entire information update process
very transparent.

As usual, compilation can be achieved with:

    $ GOOS=linux GOARCH=arm go build

Please note this project is not currently deployed anywhere: it has been superseded by the implementation
residing on `../mb-master`.
