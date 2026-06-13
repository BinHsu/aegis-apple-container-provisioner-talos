// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package create

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clustercmd "github.com/siderolabs/talos/cmd/talosctl/cmd/mgmt/cluster"
	"github.com/siderolabs/talos/cmd/talosctl/cmd/mgmt/cluster/create/clusterops"
	"github.com/siderolabs/talos/pkg/cli"
	"github.com/siderolabs/talos/pkg/provision/providers"
)

func init() {
	aOps := clusterops.GetApple()
	cOps := clusterops.GetCommon()

	const subnetFlag = "subnet"

	getAppleFlags := func() *pflag.FlagSet {
		apple := pflag.NewFlagSet("apple-container", pflag.PanicOnError)

		apple.StringVar(&aOps.TalosImage, "image", aOps.TalosImage, "the talos image to run")

		return apple
	}

	commonFlags := getCommonUserFacingFlags(&cOps)
	commonFlags.StringVar(&cOps.NetworkCIDR, subnetFlag, cOps.NetworkCIDR, "vmnet subnet CIDR (informational; DHCP assigns node IPs)")

	createAppleCmd := &cobra.Command{
		Use:   "apple-container",
		Short: "Create a local Apple container based Kubernetes cluster",
		Long: `Provisions Talos nodes as Apple container micro-VMs (one micro-VM per node).

Node IPs are assigned by vmnet DHCP and discovered at runtime; the provider applies each node's
configuration over the maintenance API after launch (there is no static-IP assignment in
apple/container), then talosctl bootstraps and waits for the cluster to be healthy.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.WithContext(context.Background(), func(ctx context.Context) error {
				provisioner, err := providers.Factory(ctx, providers.AppleContainerProviderName)
				if err != nil {
					return err
				}

				clusterConfigs, err := getAppleClusterRequest(cOps, aOps, provisioner)
				if err != nil {
					return err
				}

				cluster, err := provisioner.Create(ctx, clusterConfigs.ClusterRequest, clusterConfigs.ProvisionOptions...)
				if err != nil {
					return err
				}

				if err = postCreate(ctx, cOps, cluster, clusterConfigs); err != nil {
					return err
				}

				return clustercmd.ShowCluster(cluster)
			})
		},
	}

	createAppleCmd.Flags().AddFlagSet(getAppleFlags())
	createAppleCmd.Flags().AddFlagSet(commonFlags)

	createCmd.AddCommand(createAppleCmd)
}
