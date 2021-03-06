package helpers

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	"github.com/pivotal-cf-experimental/destiny/etcdwithops"
	"github.com/pivotal-cf-experimental/destiny/ops"
)

func NewEtcdManifestWithOpsWithInstanceCountAndReleaseVersion(deploymentPrefix string, instanceCount int, enableSSL bool, boshClient bosh.Client, releaseVersion string) (string, error) {
	manifestName := fmt.Sprintf("etcd-%s", deploymentPrefix)

	//TODO: AZs should be pulled from integration_config
	var (
		manifest string
		err      error
	)
	manifest, err = etcdwithops.NewManifestV2(etcdwithops.ConfigV2{
		Name:      manifestName,
		AZs:       []string{"z1", "z2"},
		EnableSSL: enableSSL,
	})

	if err != nil {
		return "", err
	}

	manifest, err = ops.ApplyOp(manifest, ops.Op{
		Type:  "replace",
		Path:  "/releases/name=etcd/version",
		Value: releaseVersion,
	})
	if err != nil {
		return "", err
	}

	manifest, err = ops.ApplyOp(manifest, ops.Op{
		Type:  "replace",
		Path:  "/instance_groups/name=etcd/instances",
		Value: instanceCount,
	})
	if err != nil {
		return "", err
	}

	manifestYAML, err := boshClient.ResolveManifestVersionsV2([]byte(manifest))
	if err != nil {
		return "", err
	}

	return string(manifestYAML), nil
}

func NewEtcdManifestWithOpsWithInstanceCount(deploymentPrefix string, instanceCount int, enableSSL bool, boshClient bosh.Client) (string, error) {
	return NewEtcdManifestWithOpsWithInstanceCountAndReleaseVersion(deploymentPrefix, instanceCount, enableSSL, boshClient, EtcdDevReleaseVersion())
}

func DeployEtcdWithOpsWithInstanceCountAndReleaseVersion(deploymentPrefix string, instanceCount int, enableSSL bool, boshClient bosh.Client, releaseVersion string) (string, error) {
	manifest, err := NewEtcdManifestWithOpsWithInstanceCountAndReleaseVersion(deploymentPrefix, instanceCount, enableSSL, boshClient, releaseVersion)
	if err != nil {
		return "", err
	}

	_, err = boshClient.Deploy([]byte(manifest))
	if err != nil {
		return "", err
	}

	return manifest, nil
}

func DeployEtcdWithOpsWithInstanceCount(deploymentPrefix string, instanceCount int, enableSSL bool, boshClient bosh.Client) (string, error) {
	return DeployEtcdWithOpsWithInstanceCountAndReleaseVersion(deploymentPrefix, instanceCount, enableSSL, boshClient, EtcdDevReleaseVersion())
}
