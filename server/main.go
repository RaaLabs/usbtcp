package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/pkg/term"
	"go.bug.st/serial/enumerator"
)

type netConfig struct {
	listenIPPort string

	mtls   bool
	caCert string
	cert   string
	key    string
}

func main() {
	vid := flag.String("vid", "", "usb VID")
	pid := flag.String("pid", "", "usb PID")
	mtls := flag.Bool("mtls", false, "set to true to enable, and also set caCert and cert flags")
	caCert := flag.String("caCert", "../certs/ca-cert.pem", "the path to the ca certificate. There is a helper script 'gencert.sh' who will generate self signed certificates if you don't have other certificates to use")
	cert := flag.String("cert", "../certs/server-cert.pem", "the path to the server certificate")
	key := flag.String("key", "../certs/server-key.pem", "the path to the private key")
	flag.Parse()

	nConf := netConfig{
		listenIPPort: "127.0.0.1:45000",
		mtls:         *mtls,
		caCert:       *caCert,
		cert:         *cert,
		key:          *key,
	}

	ttyName, err := getTTY(*vid, *pid)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	fmt.Printf("info: found port: %v\n", ttyName)

	err = relay(ttyName, nConf)
	if err != nil {
		log.Printf("%v\n", err)
	}
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
		conn, err := nl.Accept()
		if err != nil {
			log.Printf("error: opening out endpoint failed: %v\n", err)
			continue
		}

		// Read tty -> write net.Conn
		go func() {
			for {
				b := make([]byte, 1)
				_, err := tty.Read(b)
				if err != nil && err != io.EOF {
					log.Printf("error: fh, failed to read : %v\n", err)
					return
				}

				// fmt.Printf(" * reading tty string: %v, characters: %v\n", string(b), n)

				_, err = conn.Write(b)
				if err != nil {
					log.Printf("error: pt.Write: %v\n", err)
					return
				}

				//fmt.Printf("wrote to conn: %v\n", string(b))

			}
		}()

		// Read net.Conn -> write tty
		for {
			b := make([]byte, 1)

			_, err := conn.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("error: failed to read pt : %v\n", err)
				continue
			}
			if err == io.EOF {
				return fmt.Errorf("error: pt.Read, got io.EOF: %v", err)
			}

			// fmt.Printf(" * reading conn string: %v, characters: %v\n", string(b), n)

			_, err = tty.Write(b)
			if err != nil {
				return fmt.Errorf("error: fh.Write : %v", err)
			}

			// fmt.Printf("wrote %v charachters to fh: %s\n", n, b)
		}

	}
}

func getNetListener(nConf netConfig) (net.Listener, error) {
	switch nConf.mtls {
	case true:
		log.Printf("info: loading certificate\n")

		cert, err := tls.LoadX509KeyPair(nConf.cert, nConf.key)
		if err != nil {
			return nil, fmt.Errorf("error: failed to open cert: %v", err)
		}

		certPool := x509.NewCertPool()
		pemCABytes, err := ioutil.ReadFile(nConf.caCert)
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

		nl, err := tls.Listen("tcp", nConf.listenIPPort, config)
		if err != nil {
			return nil, fmt.Errorf("error: failed to start server listener: %v", err)
		}

		log.Printf("info: done loading certificate\n")

		return nl, nil

	case false:
		nl, err := net.Listen("tcp", nConf.listenIPPort)
		if err != nil {
			return nl, fmt.Errorf("error: opening network listener failed: %v", err)
		}

		return nl, nil
	}

	return nil, fmt.Errorf("error: opening network listener failed: unable to get state of mtls flag")
}
