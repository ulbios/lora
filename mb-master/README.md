# Transparent Modbus/TCP Slave
This directory contains an implementation superseding all those found on `../mb-{emitter,gateway,server}`.
Instead of directly interfacing with a LoRa-capable radio (i.e. the RFM9x) we instead chose to leverage
a commercial Modbus-to-LoRa bridge converter. This piece of equipment works in such a way that any incoming
Modbus RTU frames are radiated over LoRa **as long as** the destination Slave ID is not `1` (i.e. the ID
of the bridge itself). In fact, Slave ID `1` is leveraged by the bridge itself for its configuration. This
scheme makes the querying of remote Modbus-enabled devices over LoRa transparent: one just needs to query
a Slave ID other than `1` and the RTU frame will be radiated for other bridges to pick it up.

Queried data will be inserted into memory area serving as the Modbus Slave's registers. This will be queried
over TCP by a remote industrial-grade PLC implementing a Modbus/TCP client. The register addresses where
data is served is fully configurable through command-line flags.

This implementation is **currently deployed** on the water quality control hub at Guadalajara's EDAR on
a Raspberry Pi 4 reachable on IPv4 address `192.168.240.4`. The implementation is running headlessly
as a SystemD unit defined on `mb-master.service`. The applicable and interesting options are explicitly
defined on the `ExecStart` clause, which would make the appropriate invocation:

    /bin/mb-master --local-mb-bind-address 0.0.0.0 --local-mb-bind-port 502           \
        --remote-mb-enable --remote-mb-serial-dev /dev/ttyUSB0 --remote-mb-timeout 60 \
        --udp-enable --udp-bind-address 0.0.0.0

## Compilation
In order to ease compilation, this implementation includes a `Taskfile` offering the following targets:

- `mac`: Builds the application for macOS machines.
- `arm`: Builds the application for ARM devices such as the Raspberry Pi 4.

Running any target is a mater of invoking `task <target-name>`.

## Debugging
In order to test data insertion we included a piece of code exposing a UDP socket. We leaned towards UDP
because it is much easier to manage than TCP and it can be leveraged with tools such as `nc(1)`. Once
can invoke the application and then insert data with:

    # Run the server on another terminal: we'll bind the UDP socket to 127.0.0.1:1503
    $  ./bin/mb-master-mac --remote-mb-enable=false

    # And then fire up netcat (in UDP mode): you can type stuff on the session that opens up.
    $ nc -u 127.0.0.1 1503
    {"deviceId": "geonosis", "data": 5}

The data insertion will be reflected on the logs! On top of that, one can use the `mb-query` tool
to verify data can be retrieved:

    $ mb-query tcp --host 127.0.0.1 --port 1502 readHolding <register-address>

What's more, address `0` contains static value `0xABCD`:

    $ mb-query tcp --host 127.0.0.1 --port 1502 readHolding 0
    Data for address 0x0:
        Decimal -> 43981
        Hex -> 0xabcd
        Octal -> 0125715
