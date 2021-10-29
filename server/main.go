package main

import (
	"io"
	"log"
	"net"

	"github.com/pkg/term"
)

func main() {
	// --- Server: Open tty

	tty, err := term.Open("/dev/ttyUSB0")

	if err != nil {
		log.Printf("error: tty OpenFile: %v\n", err)
	}
	defer tty.Close()
	defer tty.Restore()
	term.RawMode(tty)

	err = tty.SetSpeed(9600)
	if err != nil {
		log.Printf("error: failed to set baud: %v\n", err)
		return
	}

	// --- Server: Open network listener

	nl, err := net.Listen("tcp", "127.0.0.1:45000")
	if err != nil {
		log.Printf("error: opening network listener failed: %v\n", err)
		return
	}
	defer nl.Close()

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	for {
		conn, err := nl.Accept()
		if err != nil {
			log.Printf("error: opening out endpoint failed: %v\n", err)
			continue
		}

		// Read tty -> write net.Conn
		go func() {
			for {
				b := make([]byte, 64)
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
				log.Printf("error: pt.Read, got io.EOF: %v\n", err)
				return
			}

			// fmt.Printf(" * reading conn string: %v, characters: %v\n", string(b), n)

			_, err = tty.Write(b)
			if err != nil {
				log.Printf("error: fh.Write : %v\n", err)
				return
			}

			// fmt.Printf("wrote %v charachters to fh: %s\n", n, b)
		}

	}
}
