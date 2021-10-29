package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/term"
	"go.bug.st/serial/enumerator"
)

func main() {
	vid := flag.String("vid", "", "usb VID")
	pid := flag.String("pid", "", "usb PID")
	flag.Parse()

	ttyName, err := getTTY(*vid, *pid)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	fmt.Printf("info: found port: %v\n", ttyName)

	err = relay(ttyName)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

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

func relay(ttyName string) error {
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

	// --- Server: Open network listener

	nl, err := net.Listen("tcp", "127.0.0.1:45000")
	if err != nil {
		return fmt.Errorf("error: opening network listener failed: %v", err)
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
