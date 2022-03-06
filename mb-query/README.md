# A simple ModBus query client
We **really** need to work on this (and other) READMEs, but the most important thing is we actually remember how to leverage PTYs to feasibly test serial communication between our ModBus server and client. We got the idea for all this from [this](https://superuser.com/questions/1356412/socat-gives-resource-temporarily-unavailable-on-os-x-high-sierra) SuperUser question.

There's a rather powerful (yet, fairly complex) tool called [`socat(1)`](https://manpages.debian.org/experimental/socat/socat.1.en.html) that comes in very handy when working with byte streams in general. In our case, we'll use it instead of having to cope with the `/dev/ptmx` interface or the alternatives offered in BSD-flavoured systems. You can read a bit more on what PTYs are and such on [`pty(7)`](https://man7.org/linux/man-pages/man7/pty.7.html).

Be it what it may, we'll use `socat` to open two PTYs:

    socat PTY,link=/tmp/ptyA PTY,link=/tmp/ptyB

Under the hood, this looks a bit like:

    ModBus Client                                                     ModBus Slave
    +-----------+       +--------------+       +--------------+       +-----------+
    | /tmp/ptyB | ----> | PTY B Master | ----> | PTY A Master | ----> | /tmp/ptyA |
    +-----------+       +--------------+       +--------------+       +-----------+
      PTY Slave         ^                                     ^         PTY Slave
                         \                                   /
                          + ------------ Socat ------------ +

If we are to put the above in words we could say `socat` is 'bridging' the masters of both pseudoterminals so that whatever is received on one is sent out the other. This dynamic is pretty much what you observe between a PTY master and the slave: whatever you write on one can be read from the other. This allows us to establish a link beginning in a PTY slave. When we write to it, the data will show up on the associated master. Then, `socat` will write that on the other master, which, eventually, makes the data available on the other slave.

This approach has two major advantages:

1. We **do not** have to deal with the intricacies of setting up these PTYs: it requires some system calls and such which we were not willing to jump through to just validate our configuration.

2. We can make the code we're to try and the infrastructure allowing us to test it independent. In other words, if we were to leverage a single PTY we would almost certainly have to resort to multiprocessing functions such as `fork()` to share tha PTY's associated file descriptors among the client and the server. This would imply writing at least a bit of code to just mesh the two together, which is something we'd rather not delve into.
