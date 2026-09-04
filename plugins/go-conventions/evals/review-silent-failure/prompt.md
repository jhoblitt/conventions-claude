There is no checkout of the repository under review, no network, and no Go
toolchain in this environment, and nothing can be installed: `go`, `gopls`,
and `golangci-lint` are not on PATH, and the plugin's own tools cannot be
built or run. The diff below is the whole change — one branch against its
base — in a module whose `go.mod` says `module example.com/svc` and
`go 1.27`. The repository has no `CLAUDE.md`. `Sync` is the only caller of
`Flush`, in this package or any other, and `fallback` is unchanged. Treat the
diff as complete; there is nothing further to fetch.

Review this. Your entire final answer is the review. Nothing else.

```diff
diff --git a/internal/store/store.go b/internal/store/store.go
index 06eb1cc..6485103 100644
--- a/internal/store/store.go
+++ b/internal/store/store.go
@@ -2,6 +2,9 @@ package store
 
 import (
 	"context"
+	"errors"
+	"fmt"
+	"log/slog"
 	"os"
 )
 
@@ -11,23 +14,32 @@ type Store struct {
 	records []string
 }
 
-func (s *Store) flushLegacy() error {
+// Flush writes the pending records to the file at path.
+func (s *Store) Flush(ctx context.Context) error {
 	f, err := os.Create(s.path)
 	if err != nil {
-		return err
+		return fmt.Errorf("failed: %v", err)
 	}
-	for _, r := range s.records {
-		if _, err := f.WriteString(r); err != nil {
-			f.Close()
-			return err
+	defer func() { _ = f.Close() }()
+
+	for i := range len(s.records) {
+		if _, err := f.WriteString(s.records[i]); err != nil {
+			slog.ErrorContext(ctx, "record write failed", "index", i, "err", err)
 		}
 	}
-	return f.Close()
+
+	if err != nil {
+		return fmt.Errorf("failed: %v", err)
+	}
+	return nil
 }
 
 // Sync flushes the pending records to disk.
 func (s *Store) Sync(ctx context.Context) error {
-	if err := s.flushLegacy(); err != nil {
+	if err := s.Flush(ctx); err != nil {
+		if errors.Is(err, os.ErrPermission) {
+			return s.fallback(ctx)
+		}
 		return err
 	}
 	return nil
```
