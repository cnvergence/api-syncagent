/*
Copyright 2025 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync

import ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

const (
	// deletionFinalizer is the finalizer put on remote objects to prevent
	// them from being deleted before the local objects can be cleaned up.
	deletionFinalizer = "syncagent.kcp.io/cleanup"

	// The following 4 labels/annotations are put on local objects to link them to their
	// origin remote objects. Note that the cluster *path* label is optional and
	// has to be enabled per PublishedResource.

	remoteObjectClusterLabel       = "syncagent.kcp.io/remote-object-cluster"
	remoteObjectNamespaceHashLabel = "syncagent.kcp.io/remote-object-namespace-hash"
	remoteObjectNameHashLabel      = "syncagent.kcp.io/remote-object-name-hash"

	remoteObjectNamespaceAnnotation = "syncagent.kcp.io/remote-object-namespace"
	remoteObjectNameAnnotation      = "syncagent.kcp.io/remote-object-name"

	remoteObjectWorkspacePathAnnotation = "syncagent.kcp.io/remote-object-workspace-path"

	// agentNameLabel contains the Sync Agent's name and is used to allow multiple Sync Agents
	// on the same service cluster, syncing *the same* API to different kcp's.
	agentNameLabel = "syncagent.kcp.io/agent-name"

	// objectStateLabelName is put on object state Secrets to allow for easier mass deletions
	// if ever necessary.
	objectStateLabelName = "syncagent.kcp.io/object-state"

	// objectStateLabelValue is the value of the objectStateLabelName label.
	objectStateLabelValue = "true"

	// relatedObjectAnnotationPrefix is the prefix for the annotation that is placed on
	// objects in the kcp workspaces, informing the user about the existence of a related
	// object. The identifier of the related object is appended to this to form the
	// full annotation name, the annotation value is a JSON string containing GVK and
	// metadata of the related object.
	relatedObjectAnnotationPrefix = "related-resources.syncagent.kcp.io/"
)

func OwnedBy(obj ctrlruntimeclient.Object, agentName string) bool {
	return obj.GetLabels()[agentNameLabel] == agentName
}
