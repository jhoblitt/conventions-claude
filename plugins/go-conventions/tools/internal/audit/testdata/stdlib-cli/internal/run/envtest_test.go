//go:build integration

package run

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// The only controller-runtime in this module: a test harness, not the logger or
// the manager the canon's kubernetes rules defer to.
func TestAgainstAPIServer(t *testing.T) {
	env := &envtest.Environment{}
	if _, err := env.Start(); err != nil {
		t.Skipf("no control plane: %v", err)
	}
	defer func() { _ = env.Stop() }()
}
