package main

import (
	"log"
	"net"
)

func serve(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", ln.Addr())

	return ln.Close()
}
