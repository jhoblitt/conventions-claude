package audit_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var update = flag.Bool("update", false, "rewrite the golden files under testdata")

func TestAudit(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteCfg, reporterCfg := GinkgoConfiguration()
	suiteCfg.RandomizeAllSpecs = true
	suiteCfg.FailOnPending = true

	RunSpecs(t, "audit suite", suiteCfg, reporterCfg)
}
