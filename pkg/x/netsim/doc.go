// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package netsim provides a simple network simulation framework
that developers can use to write integration tests.

# Usage and Features

The [NewStack] function creates a new, simulated network stack
using a given IP address. You can invoke usual functions on the
stack, such as:

- DialContext
- Listen
- ListenPacket

These functions return simulated [net.Conn], [net.Listener], and
[net.PacketConn] respectively.

When a connection sends data, the data is wrapped inside a [*Packet]
emitted on the channel returned by [*Stack.Output]. The [*Link]
type allows connecting two [*Stack] such that they can send [*Packet]
to each other. To send a [*Packet] to a [*Stack], you need to post
the packet on the channel returned by [*Stack.Input]. You don't need
to use a [*Link] as long as you correctly forward packets. In fact,
for simulating complex censorship scenarios, you probably want to
write custom code to forward or drop [*Packet]. In the future, there
will be subpackages of [netsim] providing this functionality.

Subpackages of this package contain extensions. For example, the
[netsim/simpki] package code helps to simulate a PKI.

The implementation of [net.Conn], [net.Listener], and [net.PacketConn] are
[*TCPConn], [*UDPConn], and [*UDPListener]. These types, which can also
be created manually, are tiny wrappers around [*Port], which contains most
of the common implementation code. These types are public to enable writing
more complex tests (e.g., the sending of unexpected TCP flags).

The errors returned by these types are the same [syscall.Errno] the
standard library and the kernel would generate in similar cases (we use
the [x/sys] repository to pull system-dependent error values).

This package contains comprehensive examples showing how to use it.

# Design Documents

This package is experimental and has no design documents for now.
*/
package netsim
