package helpers

import "github.com/cloudfoundry/bosh-bootloader/bosh"

func DeployEtcdV2WithInstanceCount(deploymentPrefix string, count int, client bosh.Client, config Config) (etcd.ManifestV2, error) {
	manifest, err = NewEtcdV2WithInstanceCount(deploymentPrefix, count, client, config, enableSSL)
	if err != nil {
		return etcd.ManifestV2{}, err
	}

	err = ResolveVersionsAndDeploy(manifest, client)
	if err != nil {
		return etcd.ManifestV2{}, err
	}

	return manifest, nil
}

func NewEtcdV2WithInstanceCount(deploymentPrefix string, count int, client bosh.Client, config Config) (etcd.ManifestV2, error) {
	manifestConfig = etcd.ConfigV2{}
	iaasConfig = etcd.IAASConfigV2{}

	manifest, err := etcd.NewManifestV2(manifestConfig, iaasConfig)
	if err != nil {
		return etcd.ManifestV2{}, err
	}

	manifest, err = etcd.SetInstanceCount(manifest, "etcd_z1", count)
	if err != nil {
		return etcd.ManifestV2{}, err
	}

	return manifest, nil
}
