package turbulence_test

import (
	etcdclient "acceptance-tests/testing/etcd"
	"acceptance-tests/testing/helpers"
	"fmt"
	"math/rand"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	turbulenceclient "github.com/pivotal-cf-experimental/bosh-test/turbulence"
	"github.com/pivotal-cf-experimental/destiny/etcd"
	"github.com/pivotal-cf-experimental/destiny/iaas"
	"github.com/pivotal-cf-experimental/destiny/turbulence"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KillVm", func() {
	KillVMTest := func(enableSSL bool, ipOffset int, turbulenceJobIPOffset int) {
		var (
			etcdManifest etcd.Manifest
			etcdClient   etcdclient.Client

			testKey1   string
			testValue1 string

			testKey2   string
			testValue2 string

			turbulenceManifest turbulence.Manifest
			turbulenceClient   turbulenceclient.Client
		)

		BeforeEach(func() {
			By("deploying turbulence", func() {
				info, err := client.Info()
				Expect(err).NotTo(HaveOccurred())

				guid, err := helpers.NewGUID()
				Expect(err).NotTo(HaveOccurred())

				manifestConfig := turbulence.Config{
					DirectorUUID: info.UUID,
					Name:         "turbulence-etcd-" + guid,
					IPOffset:     turbulenceJobIPOffset,
					BOSH: turbulence.ConfigBOSH{
						Target:         config.BOSH.Target,
						Username:       config.BOSH.Username,
						Password:       config.BOSH.Password,
						DirectorCACert: config.BOSH.DirectorCACert,
					},
				}

				var iaasConfig iaas.Config
				switch info.CPI {
				case "aws_cpi":
					if config.AWS.Subnet == "" {
						Fail("aws.subnet is required for AWS IAAS deployment")
					}

					manifestConfig.IPRange = "10.0.16.0/24"
					iaasConfig = iaas.AWSConfig{
						AccessKeyID:           config.AWS.AccessKeyID,
						SecretAccessKey:       config.AWS.SecretAccessKey,
						DefaultKeyName:        config.AWS.DefaultKeyName,
						DefaultSecurityGroups: config.AWS.DefaultSecurityGroups,
						Region:                config.AWS.Region,
						Subnet:                config.AWS.Subnet,
						RegistryHost:          config.Registry.Host,
						RegistryPassword:      config.Registry.Password,
						RegistryPort:          config.Registry.Port,
						RegistryUsername:      config.Registry.Username,
					}
				case "warden_cpi":
					iaasConfig = iaas.NewWardenConfig()
					manifestConfig.IPRange = "10.244.16.0/24"
				default:
					Fail("unknown infrastructure type")
				}

				turbulenceManifest = turbulence.NewManifest(manifestConfig, iaasConfig)

				yaml, err := turbulenceManifest.ToYAML()
				Expect(err).NotTo(HaveOccurred())

				yaml, err = client.ResolveManifestVersions(yaml)
				Expect(err).NotTo(HaveOccurred())

				fmt.Println(string(yaml))

				turbulenceManifest, err = turbulence.FromYAML(yaml)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.Deploy(yaml)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() ([]bosh.VM, error) {
					return client.DeploymentVMs(turbulenceManifest.Name)
				}, "1m", "10s").Should(ConsistOf([]bosh.VM{
					{Index: 0, JobName: "api", State: "running"},
				}))
			})

			By("preparing turbulence client", func() {
				turbulenceUrl := fmt.Sprintf("https://turbulence:%s@%s:8080",
					turbulenceManifest.Properties.TurbulenceAPI.Password,
					turbulenceManifest.Jobs[0].Networks[0].StaticIPs[0])

				turbulenceClient = turbulenceclient.NewClient(turbulenceUrl, 5*time.Minute, 2*time.Second)
			})
		})

		AfterEach(func() {
			By("deleting the turbulence deployment", func() {
				if !CurrentGinkgoTestDescription().Failed {
					err := client.DeleteDeployment(turbulenceManifest.Name)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})

		BeforeEach(func() {
			guid, err := helpers.NewGUID()
			Expect(err).NotTo(HaveOccurred())

			testKey1 = "etcd-key-1-" + guid
			testValue1 = "etcd-value-1-" + guid

			testKey2 = "etcd-key-2-" + guid
			testValue2 = "etcd-value-2-" + guid

			etcdManifest, err = helpers.DeployEtcdWithInstanceCount(3, client, config, enableSSL, ipOffset)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]bosh.VM, error) {
				return client.DeploymentVMs(etcdManifest.Name)
			}, "1m", "10s").Should(ConsistOf(helpers.GetVMsFromManifest(etcdManifest)))

			etcdClient = etcdclient.NewClient(fmt.Sprintf("http://%s:6769", etcdManifest.Jobs[2].Networks[0].StaticIPs[0]))
		})

		AfterEach(func() {
			By("deleting the deployment", func() {
				if !CurrentGinkgoTestDescription().Failed {
					err := client.DeleteDeployment(etcdManifest.Name)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})

		Context("when a etcd node is killed", func() {
			It("is still able to function on healthy vms and recover", func() {
				By("setting a persistent value", func() {
					err := etcdClient.Set(testKey1, testValue1)
					Expect(err).ToNot(HaveOccurred())
				})

				By("killing indices", func() {
					err := turbulenceClient.KillIndices(etcdManifest.Name, "etcd_z1", []int{rand.Intn(3)})
					Expect(err).ToNot(HaveOccurred())
				})

				By("reading the value from etcd", func() {
					value, err := etcdClient.Get(testKey1)
					Expect(err).ToNot(HaveOccurred())
					Expect(value).To(Equal(testValue1))
				})

				By("setting a new persistent value", func() {
					err := etcdClient.Set(testKey2, testValue2)
					Expect(err).ToNot(HaveOccurred())
				})

				By("fixing the deployment", func() {
					yaml, err := etcdManifest.ToYAML()
					Expect(err).NotTo(HaveOccurred())

					err = client.ScanAndFix(yaml)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func() ([]bosh.VM, error) {
						return client.DeploymentVMs(etcdManifest.Name)
					}, "1m", "10s").Should(ConsistOf(helpers.GetVMsFromManifest(etcdManifest)))
				})

				By("reading each value from the resurrected VM", func() {
					value, err := etcdClient.Get(testKey1)
					Expect(err).ToNot(HaveOccurred())
					Expect(value).To(Equal(testValue1))

					value, err = etcdClient.Get(testKey2)
					Expect(err).ToNot(HaveOccurred())
					Expect(value).To(Equal(testValue2))
				})
			})
		})
	}

	Context("without TLS", func() {
		KillVMTest(false, helpers.KillVMWithoutTLSIPOffset, helpers.KillVMWithoutTLSTurbulenceJobIPOffset)
	})

	Context("with TLS", func() {
		KillVMTest(true, helpers.KillVMWithTLSIPOffset, helpers.KillVMWithTLSTurbulenceJobIPOffset)
	})
})
