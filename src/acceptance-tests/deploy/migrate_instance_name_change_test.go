package deploy_test

import (
	"fmt"
	"time"

	"github.com/greenhouse-org/consul-release/src/acceptance-tests/testing/helpers"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	"github.com/pivotal-cf-experimental/destiny/etcd"
)

var _ = FDescribe("Migrate instance name change", func() {
	var (
		manifest   etcd.Manifest
		etcdClient etcdclient.Client
		spammer    *helpers.Spammer

		testKey   string
		testValue string
	)

	BeforeEach(func() {
		guid, err := helpers.NewGUID()
		Expect(err).NotTo(HaveOccurred())

		testKey = "etcd-key-" + guid
		testValue = "etcd-value-" + guid

		manifestV2, err = helpers.DeployEtcdV2WithInstanceCount("migrate_instance_name_change", 3, client, config, true)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() ([]bosh.VM, error) {
			return helpers.DeploymentVMs(client, manifest.Name)
		}, "1m", "10s").Should(ConsistOf(helpers.GetVMsFromManifest(manifest)))

		testConsumerIndex, err := helpers.FindJobIndexByName(manifest, "testconsumer_z1")
		Expect(err).NotTo(HaveOccurred())
		etcdClient = etcdclient.NewClient(fmt.Sprintf("http://%s:6769", manifest.Jobs[testConsumerIndex].Networks[0].StaticIPs[0]))
		spammer = helpers.NewSpammer(etcdClient, 1*time.Second, "migrate-instance-name-change")
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			err := client.DeleteDeployment(manifest.Name)
			Expect(err).NotTo(HaveOccurred())
		}
	})
	It("migrates an etcd cluster sucessfully when instance name is changed", func() {
		By("setting a persistent value", func() {
			err := etcdClient.Set(testKey, testValue)
			Expect(err).ToNot(HaveOccurred())
		})

		By("deploying with a new name", func() {

		})

	})
})
