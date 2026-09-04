package main

import (
	"flag"
	"log"
	"os"

	"example.com/x/internal/run"
)

func main() {
	name := flag.String("name", "world", "who to greet")
	flag.Parse()

	if err := run.Greet(os.Stdout, *name); err != nil {
		log.Fatalf("greet: %v", err)
	}
}
