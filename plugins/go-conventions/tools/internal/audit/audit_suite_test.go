package audit_test

import (
	"flag"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var update = flag.Bool("update", false, "rewrite the golden files under testdata")

// pluginRoot is the go-conventions plugin directory this module ships inside,
// the value CLAUDE_PLUGIN_ROOT carries in production. The specs measure against
// the real templates rather than a copy, so a template change moves the
// goldens.
var pluginRoot string

func TestAudit(t *testing.T) {
	RegisterFailHandler(Fail)

	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving the plugin root: %v", err)
	}
	pluginRoot = root

	suiteCfg, reporterCfg := GinkgoConfiguration()
	suiteCfg.RandomizeAllSpecs = true
	suiteCfg.FailOnPending = true

	RunSpecs(t, "audit suite", suiteCfg, reporterCfg)
}
