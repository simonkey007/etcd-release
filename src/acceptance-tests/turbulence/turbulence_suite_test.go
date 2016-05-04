package turbulence_test

import (
	"acceptance-tests/testing/helpers"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTurbulence(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "turbulence")
}

var (
	config helpers.Config
	client bosh.Client
)

var _ = BeforeSuite(func() {
	configPath, err := helpers.ConfigPath()
	Expect(err).NotTo(HaveOccurred())

	config, err = helpers.LoadConfig(configPath)
	Expect(err).NotTo(HaveOccurred())

	client = bosh.NewClient(bosh.Config{
		URL:              fmt.Sprintf("https://%s:25555", config.BOSH.Target),
		Username:         config.BOSH.Username,
		Password:         config.BOSH.Password,
		AllowInsecureSSL: true,
	})
})
