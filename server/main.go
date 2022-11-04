package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/term"
	"go.bug.st/serial/enumerator"
)

type netConfig struct {
	mtls   bool
	caCert string
	cert   string
	key    string

	baud           int
	ipPort         string
	ttyReadTimeout int
}

// getTTY will get the path of the tty.
func getTTY(vid string, pid string) (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		return "", fmt.Errorf("error: no serial port found with that ID")
	}
	for _, port := range ports {
		if port.IsUSB {
			if port.VID == vid && port.PID == pid {
				return port.Name, nil
			}
		}
	}

	return "", fmt.Errorf("error: no port with that ID found")
}

// relay will start relaying the data between the TTY and the network connection.
func relay(ttyName string, nConf netConfig) error {

	for {
		err := func() error {

			// --- Server: Open tty

			tty, err := term.Open(ttyName)

			if err != nil {
				log.Printf("error: tty OpenFile: %v\n", err)
			}
			defer tty.Close()
			defer tty.Restore()
			term.RawMode(tty)

			err = tty.SetSpeed(9600)
			if err != nil {
				return fmt.Errorf("error: failed to set baud: %v", err)
			}

			nl, err := getNetListener(nConf)
			if err != nil {
				return fmt.Errorf("error: opening network listener failed: %v", err)
			}
			defer nl.Close()

			for {

				err := tty.SetReadTimeout(time.Second * time.Duration(nConf.ttyReadTimeout))
				if err != nil {
					return fmt.Errorf("error: setReadTimeoutFailed: %v", err)
				}

				conn, err := nl.Accept()
				if err != nil {
					log.Printf("error: opening out endpoint failed: %v\n", err)
					continue
				}

				connOK := true

				errCh := make(chan error)

				// Read tty -> write net.Conn
				go func() {
					fmt.Printf(" * starting go routine for Read tty -> write net.Conn\n")
					defer fmt.Printf(" ** ending go routine for Read tty -> write net.Conn\n")

					for {

						b := make([]byte, 1)
						n, err := tty.Read(b)
						if err != nil {
							if connOK {
								fmt.Printf("connOK = %v\n", connOK)
								continue
							}

							er := fmt.Errorf("error: tty.Read failed: %v", err)
							select {
							case errCh <- er:
							default:
								fmt.Printf("%v\n", er)
							}

							return
						}

						fmt.Printf(" tty read nr = %v\n", n)

						_, err = conn.Write(b)
						if err != nil {
							errCh <- fmt.Errorf("error: conn.Write failed: %v", err)
							return
						}
					}
				}()

				// Read net.Conn -> write tty
				go func() {
					fmt.Printf(" * starting go routine for Read net.Conn -> write tty\n")
					defer fmt.Printf(" ** ending go routine for Read net.Conn -> write tty\n")
					defer func() { connOK = false }()

					for {
						b := make([]byte, 1)

						_, err := conn.Read(b)
						if err != nil && err != io.EOF {
							errCh <- fmt.Errorf("error: conn.Read failed : %v", err)
							return
						}
						if err == io.EOF {
							errCh <- fmt.Errorf("error: conn.Read failed, got io.EOF: %v", err)
							return
						}

						_, err = tty.Write(b)
						if err != nil {
							er := fmt.Errorf("error: tty.Write failed : %v", err)
							select {
							case errCh <- er:
							default:
								fmt.Printf("%v\n", er)
							}
							return
						}
					}
				}()

				err = <-errCh
				if err != nil {
					log.Printf("%v\n", err)
				}
				tty.Close()
				conn.Close()

			}
		}()

		if err != nil {
			log.Printf("%v\n", err)
		}
	}

}

// getNetListener will return either an normal or TLS encryptet net.Listener.
func getNetListener(nConf netConfig) (net.Listener, error) {
	switch nConf.mtls {
	case true:
		log.Printf("info: loading certificate\n")

		cert, err := tls.LoadX509KeyPair(nConf.cert, nConf.key)
		if err != nil {
			return nil, fmt.Errorf("error: failed to open cert: %v", err)
		}

		certPool := x509.NewCertPool()
		pemCABytes, err := os.ReadFile(nConf.caCert)
		if err != nil {
			return nil, fmt.Errorf("error: failed to read ca cert: %v", err)
		}

		if !certPool.AppendCertsFromPEM(pemCABytes) {
			return nil, fmt.Errorf("error: failed to append ca to cert pool")
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}

		nl, err := tls.Listen("tcp", nConf.ipPort, config)
		if err != nil {
			return nil, fmt.Errorf("error: failed to start server listener: %v", err)
		}

		log.Printf("info: done loading certificate\n")

		return nl, nil

	case false:
		nl, err := net.Listen("tcp", nConf.ipPort)
		if err != nil {
			return nl, fmt.Errorf("error: opening network listener failed: %v", err)
		}

		return nl, nil
	}

	return nil, fmt.Errorf("error: opening network listener failed: unable to get state of mtls flag")
}

func main() {
	vid := flag.String("vid", "", "usb VID")
	pid := flag.String("pid", "", "usb PID")
	mtls := flag.Bool("mtls", false, "set to true to enable, and also set caCert and cert flags")
	caCert := flag.String("caCert", "../certs/ca-cert.pem", "the path to the ca certificate. There is a helper script 'gencert.sh' who will generate self signed certificates if you don't have other certificates to use")
	cert := flag.String("cert", "../certs/server-cert.pem", "the path to the server certificate")
	key := flag.String("key", "../certs/server-key.pem", "the path to the private key")
	baud := flag.Int("baud", 9600, "baud rate")
	ipPort := flag.String("ipPort", "127.0.0.1:45000", "ip:port for where to start the network listener")
	ttyReadTimeout := flag.Int("ttyReadTimeout", 1, "The timeout for TTY read given in seconds")

	flag.Parse()

	nConf := netConfig{
		mtls:           *mtls,
		caCert:         *caCert,
		cert:           *cert,
		key:            *key,
		baud:           *baud,
		ipPort:         *ipPort,
		ttyReadTimeout: *ttyReadTimeout,
	}

	ttyName, err := getTTY(*vid, *pid)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Printf("info: found port: %v\n", ttyName)

	err = relay(ttyName, nConf)
	if err != nil {
		log.Printf("%v\n", err)
	}
}
