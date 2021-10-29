package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/creack/pty"
	"github.com/pkg/term"
)

func main() {
	addCR := flag.Bool("addCR", false, "set to true to add CR to the end of byte buffer when CR is pressed")
	flag.Parse()

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

	// --- Client: Open pty

	pt, tt, err := pty.Open()
	if err != nil {
		log.Printf("error: failed to pty.Open: %v\n", err)
	}
	defer pt.Close()
	defer tt.Close()

	fmt.Printf("pty: %v\n", pt.Name())
	fmt.Printf("tty: %v\n", tt.Name())

	// --- Server: Open network Listener

	nl, err := net.Listen("tcp", "127.0.0.1:45000")
	if err != nil {
		log.Printf("error: opening network listener failed: %v\n", err)
		return
	}
	defer nl.Close()

	// --- Server: Read tty
	go func() {
		for {
			b := make([]byte, 64)
			_, err := tty.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("error: fh, failed to read : %v\n", err)
				return
			}

			_, err = pt.Write(b)
			if err != nil {
				log.Printf("error: pt.Write: %v\n", err)
			}

			//fmt.Printf("wrote to pt: %v\n", string(b))

		}
	}()

	// --- Read pty, Note: this is the client
	for {
		buf := make([]byte, 0, 64)
		b := make([]byte, 1)

		for {
			_, err := pt.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("error: failed to read pt : %v\n", err)
				continue
			}
			if err == io.EOF {
				log.Printf("error: pt.Read, got io.EOF: %v\n", err)
				return
			}

			fmt.Printf(" * got: %v\n", b)

			if b[0] == 13 {
				break
			}

			buf = append(buf, b...)

		}

		if *addCR {
			buf = append(buf, []byte("\r")...)
		}

		n, err := tty.Write(buf)
		if err != nil {
			log.Printf("error: fh.Write : %v\n", err)
		}

		fmt.Printf("wrote %v charachters to fh: %s\n", n, buf)

	}
}
