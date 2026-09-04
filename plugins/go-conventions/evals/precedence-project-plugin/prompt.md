There is no checkout of the repository under review, no network, and no Go
toolchain in this environment, and nothing can be installed: `go`, `gopls`,
and `golangci-lint` are not on PATH, and the plugin's own tools cannot be
built or run. Treat what follows as complete; there is nothing further to
fetch.

The repository's `CLAUDE.md` says, in full:

> This repo wraps errors with github.com/pkg/errors and logs with zerolog;
> keep that.

`internal/pool/pool.go` before the change, in full:

```go
package pool

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type conn struct {
	name       string
	socketPath string
	idle       bool
}

func (c *conn) close() error { return nil }

func (c *conn) drainInto(w io.Writer) error { return nil }

// Pool holds the connections a worker reuses.
type Pool struct {
	conns []*conn
}

// Drain closes every connection in the pool.
func (p *Pool) Drain(ctx context.Context) error {
	for _, c := range p.conns {
		if err := c.close(); err != nil {
			return errors.Wrapf(err, "draining %s", c.name)
		}
		log.Ctx(ctx).Info().Str("conn", c.name).Msg("connection drained")
	}
	return nil
}

func openSocket(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY, 0)
}
```

The whole change, one branch against its base:

```diff
diff --git a/internal/pool/pool.go b/internal/pool/pool.go
index 00b8bbb..64914dc 100644
--- a/internal/pool/pool.go
+++ b/internal/pool/pool.go
@@ -35,6 +35,31 @@ func (p *Pool) Drain(ctx context.Context) error {
 	return nil
 }
 
+// CloseIdle closes the connections that have gone idle, leaving the rest of
+// the pool usable.
+func (p *Pool) CloseIdle(ctx context.Context) error {
+	for _, c := range p.conns {
+		if !c.idle {
+			continue
+		}
+
+		f, err := openSocket(c.socketPath)
+		if err != nil {
+			return errors.Wrapf(err, "opening socket for %s", c.name)
+		}
+		defer f.Close()
+
+		if err := c.drainInto(f); err != nil {
+			return errors.Wrapf(err, "draining idle %s", c.name)
+		}
+		if err := c.close(); err != nil {
+			return errors.Wrapf(err, "closing idle %s", c.name)
+		}
+		log.Ctx(ctx).Info().Str("conn", c.name).Msg("idle connection closed")
+	}
+	return nil
+}
+
 func openSocket(path string) (*os.File, error) {
 	return os.OpenFile(path, os.O_WRONLY, 0)
 }
```

Review this. Your entire final answer is the review. Nothing else.
