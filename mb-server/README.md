# Implementing a ModBus Slave
We need to relay data to an industrial-grade PLC which leverages the *ModBus/TCP* protocol. This called for the implementation of a full-fledged ModBus *slave*.

In order to be as flexible as possible we enabled communication over both *TCP* and regular serial interfaces such as *RS-485*. The result of our efforts is contained in the `server.go` file defining the server itself together with the `mb-server.service` unit file which allows us to run it as a detached daemon under normal circumstances.

This slave will receive data it needs to offer to teh PLC behind it. In doing so we'll have to enable some sort of data entry procedures. On the filed we'll be receiving data over a LoRa radio with a protocol we've developed ourselves. The thing is, we actually have to test stuff out beforehand! That's why we decided to enable the reception of data over UDP too. We chose *UDP* over *TCP* due to its enhanced simplicity: there's no need to manage connections and the like! Given Go's native concurrency, this all works out of the box in a parallel fashion: that's quite something!

In order to leverage our UDP data path we need to somehow send information. We can thankfully do su with `nc(1)`. We just need to tell it to use UDP instead of TCP (which is the default) and we should be good to go:

    # Run the server on another terminal: we'll bind the UDP socket to 127.0.0.1:1503
    collado@hoth:0:~$  ./bin/mb-server -serial-device none -udp-bind-address 127.0.0.1 -udp-bind-port 1503

    # And then fire up netcat: you can type stuff on the session that opens up.
    collado@hoth:0:~$ nc -u 127.0.0.1 1503
    {"msg": "A random message xD", "data": 5}

With that you should be able to then query the server to get back the information it's updated. Nevertheless, the server's log messages should provide very verbosy information making the entire information update process very transparent.
