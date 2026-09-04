There is no checkout of the repository below, no network, and no Go toolchain
in this environment, and nothing can be installed: `go`, `golangci-lint`,
`pinact` and `actionlint` are not on PATH, and the plugin's own tools cannot be
built or run. The listing below is the whole repository — every file it has.
Treat it as complete; there is nothing further to fetch.

What would converging this repository to our Go conventions change, and in what
order? Your entire final answer is that plan: what changes, which phase each
change belongs to, what has to wait on me, and anything you would normally have
checked here but could not. Nothing else.

`go.mod`

```gomod
module example.com/svc

go 1.25.0

toolchain go1.25.3

require github.com/stretchr/testify v1.11.1
```

`main.go`

```go
package main

import (
	"flag"
	"log"
	"net/http"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	log.Printf("svc %s listening on %s", version, *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
```

`main_test.go`

```go
package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionDefault(t *testing.T) {
	require.Equal(t, "dev", version)
}
```

`Makefile`

```make
build:
	go build -ldflags "-X main.version=$(shell git describe --tags)" -o bin/svc .

test:
	go test ./...
```

There is no `.golangci.yml`, no `.github/` directory, no `.goreleaser.yaml`, no
`CLAUDE.md`, and no `.gitignore`.
