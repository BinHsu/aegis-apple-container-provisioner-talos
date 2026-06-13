// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"errors"

	"github.com/siderolabs/talos/pkg/provision"
)

// result implements provision.Cluster. Mirrors the docker provider's result type.
type result struct {
	clusterInfo provision.ClusterInfo

	statePath string
}

func (res *result) Provisioner() string {
	return ProviderName
}

func (res *result) Info() provision.ClusterInfo {
	return res.clusterInfo
}

func (res *result) StatePath() (string, error) {
	if res.statePath == "" {
		return "", errors.New("state path is not set")
	}

	return res.statePath, nil
}
