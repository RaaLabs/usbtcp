# usbtcp

Relay character devices over TCP from one node to another.

## How to build

```bash
git clone https://github.com/RaaLabs/usbtcp.git
pushd usbtcp/server
go build
popd
pushd usbtcp/client
go build
popd
```

The binaries will be located under `./usbtcp/server` and `./usbtcp/client`.


## How to run

### Server

`./server -vid="0557" -pid="2008"`

or with mTLS

`./server -vid="0557" -pid="2008" -mtls=true`

### Client

`./client`

The reading of the character device is being done in RAW mode, so if you want to do testing via a minicom or hyperterminal you can use the `-addCR=true` flag to add carriage return to the payload sent to the server.

`./client -addCR=true`

Enable mTLS

`./client -addCR=true -mtls=true`

### mTLS

For authenticating and encryption mTLS are supported.

In the certs folder of the repository there is a `gencert.sh` script that will generate both the server and the client certificatates to needed, or you can use your own certificate.

You should edit the script and the `.cnf` files with information suited for your own needs.
