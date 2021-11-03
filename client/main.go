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

	"github.com/creack/pty"
)

type netConfig struct {
	listenIPPort string

	mtls   bool
	caCert string
	cert   string
	key    string
}

// newTLSConfig Will load all PEM encoded certificate from their file paths,
// and return a *tls.Config.
func newTLSConfig(nc netConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(nc.cert, nc.key)
	if err != nil {
		return nil, fmt.Errorf("error: failed to open cert: %v", err)
	}

	certPool := x509.NewCertPool()
	pemCABytes, err := ioutil.ReadFile(nc.caCert)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read ca cert: %v", err)
	}

	if !certPool.AppendCertsFromPEM(pemCABytes) {
		return nil, fmt.Errorf("error: failed to append ca to cert pool")
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return &config, nil
}

func main() {
	addCR := flag.Bool("addCR", false, "set to true to add CR to the end of byte buffer when CR is pressed")
	mtls := flag.Bool("mtls", false, "set to true to enable, and also set caCert and cert flags")
	caCert := flag.String("caCert", "../certs/ca-cert.pem", "the path to the ca certificate. There is a helper script 'gencert.sh' who will generate self signed certificates if you don't have other certificates to use")
	cert := flag.String("cert", "../certs/client-cert.pem", "the path to the server certificate")
	key := flag.String("key", "../certs/client-key.pem", "the path to the private key")
	flag.Parse()

	nConf := netConfig{
		listenIPPort: "127.0.0.1:45000",
		mtls:         *mtls,
		caCert:       *caCert,
		cert:         *cert,
		key:          *key,
	}

	// --- Client: Open pty

	pt, tt, err := pty.Open()
	if err != nil {
		log.Printf("error: failed to pty.Open: %v\n", err)
	}
	defer pt.Close()
	defer tt.Close()

	fmt.Printf("pty: %v\n", pt.Name())
	fmt.Printf("tty: %v\n", tt.Name())

	// --- Client: Open dial network

	var conn net.Conn

	switch nConf.mtls {
	case false:
		conn, err = net.Dial("tcp", nConf.listenIPPort)
		if err != nil {
			log.Printf("error: failed to connect : %v\n", err)
			return
		}

	case true:
		cfg, err := newTLSConfig(nConf)
		if err != nil {
			log.Printf("error: failed to create TLS config : %v\n", err)
			return
		}

		conn, err = tls.Dial("tcp", nConf.listenIPPort, cfg)
		if err != nil {
			log.Printf("error: failed to connect : %v\n", err)
			return
		}
	}

	defer conn.Close()

	if conn == nil {
		log.Printf("error: net.Conn == nil : %v\n", conn)
		return
	}

	// read network -> write pty
	go func() {
		for {
			n, err := io.Copy(pt, conn)
			if err != nil || n == 0 {
				log.Printf("error: io.Copy(pt,conn) charachers:%v, %v\n", n, err)
				return
			}
			fmt.Printf(" * io.Copy(pt,conn) copied %v number of bytes\n", n)

		}
	}()

	// read pty -> write network
	for {
		buf := make([]byte, 0, 64)

		for {
			b := make([]byte, 1)
			_, err := pt.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("error: failed to read pt : %v\n", err)
				continue
			}
			if err == io.EOF {
				log.Printf("error: pt.Read, got io.EOF: %v\n", err)
				return
			}

			// fmt.Printf(" * got: %v\n", b)

			if b[0] == 13 {
				break
			}

			buf = append(buf, b...)

		}

		if *addCR {
			buf = append(buf, []byte("\r")...)
		}

		n, err := conn.Write(buf)
		if err != nil {
			log.Printf("error: fh.Write : %v\n", err)
			return
		}

		fmt.Printf("wrote %v charachters to pty\n", n)
	}
}
