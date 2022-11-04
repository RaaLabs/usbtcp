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

Or precompiled binaries for Linux Amd64 architecture can be found in the releases.

## How to run

### Server flags

```bash
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
  -ttyReadTimeout int
    The timeout for TTY read given in seconds (default 1)
  -vid string
    usb VID
```

### Server

Use `lsusb` to get the vid and the pid for your USB device.

`./server -vid="0557" -pid="2008"`

or with mTLS

`./server -vid="0557" -pid="2008" -mtls=true`

### Client flags

```bash
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
  -portInfoFileDir string
    the directory path of where to store the port.info file (default "./")
```

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
```

### Client

All available flags can be found starting the client with the `-help` flag.

If the server is listening on the default port, and no authentication is enabled the client can be started with just using the default values.

#### Client extra options

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
