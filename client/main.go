package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/creack/pty"
)

func main() {
	addCR := flag.Bool("addCR", false, "set to true to add CR to the end of byte buffer when CR is pressed")
	flag.Parse()

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

	conn, err := net.Dial("tcp", "127.0.0.1:45000")
	if err != nil {
		log.Printf("error: failed to connect : %v\n", err)
		return
	}
	defer conn.Close()

	// net.Conn -> pty

	// Read network
	go func() {
		for {
			// b := make([]byte, 64)
			// _, err := conn.Read(b)
			// if err != nil && err != io.EOF {
			// 	log.Printf("error: failed to read conn : %v\n", err)
			// 	continue
			// }
			// if err != io.EOF {
			// 	log.Printf("error: got io.EOF: %v\n", err)
			// 	return
			// }

			n, err := io.Copy(pt, conn)
			if err != nil || n == 0 {
				log.Printf("error: io.Copy(pt,conn) charachers:%v, %v\n", n, err)
				return
			}
			fmt.Printf(" * io.Copy(pt,conn) copied %v number of bytes\n", n)

		}
	}()

	// Read pty
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
