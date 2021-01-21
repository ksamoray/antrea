// Copyright 2021 Antrea Authors
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

package ipam

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/vmware-tanzu/antrea/pkg/agent/config"
)

type CIDRTracker struct {
	nodeConfig  *config.NodeConfig
	nodesSynced cache.InformerSynced
}

func NewCIDRTracker(nodeInformer coreinformers.NodeInformer, nodeConfig *config.NodeConfig, ) *CIDRTracker {
	ct := &CIDRTracker{
		nodeConfig:		nodeConfig,
		nodesSynced: 	nodeInformer.Informer().HasSynced,
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nil,
		UpdateFunc: ct.updateNodeFunc,
		DeleteFunc: nil,
	})

	return ct
}

func (ct *CIDRTracker) updateNodeFunc(oldObj, newObj interface{}) {
	oldNode := newObj.(*v1.Node).DeepCopy()
	newNode := oldObj.(*v1.Node).DeepCopy()

	if newNode.Name == ct.nodeConfig.Name {
		if len(oldNode.Spec.PodCIDRs) != len(newNode.Spec.PodCIDRs) ||
			!reflect.DeepEqual(oldNode.Spec.PodCIDRs, newNode.Spec.PodCIDRs) {
			klog.Infof("==============> Node %s PodCIDRs changed %v", newNode.Name, newNode.Spec.PodCIDRs)
		} else {
			klog.Infof("==============> Node %s PodCIDRs NOT changed %v %v", newNode.Name, oldNode, newNode)
		}
	} else {
		klog.Infof("=================> Other node changed %v", newNode.Name)
	}
}

func (ct *CIDRTracker) Run(stopCh <-chan struct{}) {
	if !cache.WaitForNamedCacheSync("cidrtracker", stopCh, ct.nodesSynced) {
		return
	}
}
