package serve

import (
	"context"
	"net"
)

// Listen opens addr and closes it again; the daemon body is not the point here.
func Listen(ctx context.Context, addr string) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	return ln.Close()
}
