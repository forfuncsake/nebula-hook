# nebula-hook

nebula-hook is a sample program written only to exercise and demonstrate the library hook capabilities being experimented with in https://github.com/slackhq/nebula#310

## Implementation

The implementation is a contrived client/server application that uses a nebula overlay as its network transport. It is designed only for demonstration purposes and is not well tested or robust. 

The client runs a loop sending packets to the server's nebula IP and a port that is allowed through the nebula firewall (the port serves no other purpose as this time). The packet contains a single integer value.

On receipt of such a packet, the server will increment the integer by one and send the new value back to the nebula IP from which it was received.

On receipt of this reply packet, the client will produce a log message with the received value.

## Example Config

This repo includes sample configs for 2 nodes that both run on the same host. The included certificates were issued by a CA that was created only for this project, they are valid for 1 year from creation and should not be used for any purpose other than testing.

## Demo

In two different terminals, start the client and server as follows:
```
go run ./main.go -config example/server.yml -serve
```

```
go run ./main.go -config example/client.yml
```


Client output will appear like this:
```
INFO[0006] Handshake message sent                        handshake="map[stage:1 style:ix_psk0]" initiatorIndex=3538921260 remoteIndex=0 udpAddr="127.0.0.1:4002" vpnIp=192.168.100.2
INFO[0006] Handshake message received                    certName=hook-server durationNs=1905265265 fingerprint=6357c22a635b27c08ceaebe9d396aadfd1d8c2dcb6a2a653f191985b5119f74e handshake="map[stage:2 style:ix_psk0]" initiatorIndex=3538921260 remoteIndex=3538921260 responderIndex=1429205345 udpAddr="127.0.0.1:4002" vpnIp=192.168.100.2
INFO[0006] Hook payload received                         value=1
INFO[0010] Hook payload sent                             value=1
INFO[0010] Hook payload received                         value=2
INFO[0015] Hook payload sent                             value=2
INFO[0015] Hook payload received                         value=3
INFO[0020] Hook payload sent                             value=3
INFO[0020] Hook payload received                         value=4
INFO[0025] Hook payload sent                             value=4
INFO[0025] Hook payload received                         value=5
```