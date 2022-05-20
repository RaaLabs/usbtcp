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

#### Flags

```text
  -baud int
    baud rate (default 9600)
  -caCert string
    the path to the ca certificate. There is a helper script 'gencert.sh' who will generate self signed certificates if you don't have other certificates to use (default "../certs/ca-cert.pem")
  -cert string
    the path to the server certificate (default "../certs/server-cert.pem")
  -ipPort string
    ip:port for where to start the network listener (default "127.0.0.1:45000")
  -key string
    the path to the private key (default "../certs/server-key.pem")
  -mtls
    set to true to enable, and also set caCert and cert flags
  -pid string
    usb PID
  -vid string
    usb VID
````

### Client

`./client`

The reading of the character device is being done in RAW mode, so if you want to do testing via a minicom or hyperterminal you can use the `-addCR=true` flag to add carriage return to the payload sent to the server.

`./client -addCR=true`

Enable mTLS

`./client -addCR=true -mtls=true`

#### Flags

```text
  -caCert string
    the path to the ca certificate. There is a helper script 'gencert.sh' who will generate self signed certificates if you don't have other certificates to use (default "../certs/ca-cert.pem")
  -cert string
    the path to the server certificate (default "../certs/client-cert.pem")
  -ipPort string
    ip:port of the host to connec to (default "127.0.0.1:45000")
  -key string
    the path to the private key (default "../certs/client-key.pem")
  -mtls
    set to true to enable, and also set caCert and cert flags
```

### mTLS

For authenticating and encryption mTLS are supported.

In the certs folder of the repository there is a `gencert.sh` script that will generate both the server and the client certificatates to needed, or you can use your own certificate.

You should edit the script and the `.cnf` files with information suited for your own needs.
