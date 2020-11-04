// Copyright 2019 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	netutils "k8s.io/utils/net"

	"github.com/vmware-tanzu/antrea/pkg/apis"
	"github.com/vmware-tanzu/antrea/pkg/features"
)

const (
	defaultClusterCIDRs         = "172.18.0.0/16"
	defaultServiceCIDR          = "172.19.0.0/16"
	defaultNodeCIDRMaskSizeIPv4 = 24
	defaultNodeCIDRMaskSizeIPv6 = 64
)

type Options struct {
	// The path of configuration file.
	configFile string
	// The configuration object
	config *ControllerConfig
}

func newOptions() *Options {
	return &Options{
		config: &ControllerConfig{
			EnablePrometheusMetrics: true,
			SelfSignedCert:          true,
		},
	}
}

// addFlags adds flags to fs and binds them to options.
func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.configFile, "config", o.configFile, "The path to the configuration file")
}

// complete completes all the required options.
func (o *Options) complete(args []string) error {
	if len(o.configFile) > 0 {
		if err := o.loadConfigFromFile(); err != nil {
			return err
		}
	}
	o.setDefaults()
	return features.DefaultMutableFeatureGate.SetFromMap(o.config.FeatureGates)
}

// validate validates all the required options.
func (o *Options) validate(args []string) error {
	if len(args) != 0 {
		return errors.New("no positional arguments are supported")
	}

	// Validate ServiceCIDR
	_, _, err := net.ParseCIDR(o.config.ServiceCIDR)
	if err != nil {
		return fmt.Errorf("service CIDR %s is invalid", o.config.ServiceCIDR)
	}

	// Validate ClusterCIDRs
	cidrSplit := strings.Split(strings.TrimSpace(o.config.ClusterCIDRs), ",")
	_, err = netutils.ParseCIDRs(cidrSplit)
	if err != nil {
		return fmt.Errorf("cluster CIDRs %s is invalid", o.config.ClusterCIDRs)
	}

	return nil
}

func (o *Options) loadConfigFromFile() error {
	data, err := ioutil.ReadFile(o.configFile)
	if err != nil {
		return err
	}

	return yaml.UnmarshalStrict(data, &o.config)
}

func (o *Options) setDefaults() {
	if o.config.APIPort == 0 {
		o.config.APIPort = apis.AntreaControllerAPIPort
	}

	if o.config.ClusterCIDRs == "" {
		o.config.ClusterCIDRs = defaultClusterCIDRs
	}

	if o.config.ServiceCIDR == "" {
		o.config.ServiceCIDR = defaultServiceCIDR
	}

	if o.config.NodeCIDRMaskSizeIPv4 == 0 {
		o.config.NodeCIDRMaskSizeIPv4 = defaultNodeCIDRMaskSizeIPv4
	}

	if o.config.NodeCIDRMaskSizeIPv6 == 0 {
		o.config.NodeCIDRMaskSizeIPv6 = defaultNodeCIDRMaskSizeIPv6
	}
}
