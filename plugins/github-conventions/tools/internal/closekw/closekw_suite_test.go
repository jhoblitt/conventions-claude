package closekw_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCloseKW(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteCfg, reporterCfg := GinkgoConfiguration()
	suiteCfg.RandomizeAllSpecs = true
	suiteCfg.FailOnPending = true

	RunSpecs(t, "closekw suite", suiteCfg, reporterCfg)
}
