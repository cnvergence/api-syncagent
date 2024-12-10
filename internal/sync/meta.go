/*
Copyright 2024 The Kubermatic Kubernetes Platform contributors.

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

import (
	"context"

	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureLabels(obj metav1.Object, desiredLabels map[string]string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	for k, v := range desiredLabels {
		labels[k] = v
	}

	obj.SetLabels(labels)
}

func ensureAnnotations(obj metav1.Object, desiredAnnotations map[string]string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for k, v := range desiredAnnotations {
		annotations[k] = v
	}

	obj.SetAnnotations(annotations)
}

func ensureFinalizer(ctx context.Context, log *zap.SugaredLogger, client ctrlruntimeclient.Client, obj *unstructured.Unstructured, finalizer string) (updated bool, err error) {
	finalizers := sets.New[string](obj.GetFinalizers()...)
	if finalizers.Has(deletionFinalizer) {
		return false, nil
	}

	original := obj.DeepCopy()

	finalizers.Insert(deletionFinalizer)
	obj.SetFinalizers(sets.List(finalizers))

	log.Debugw("Adding finalizer…", "on", newObjectKey(obj, ""), "finalizer", finalizer)
	if err := client.Patch(ctx, obj, ctrlruntimeclient.MergeFrom(original)); err != nil {
		return false, err
	}

	return true, nil
}

func removeFinalizer(ctx context.Context, log *zap.SugaredLogger, client ctrlruntimeclient.Client, obj *unstructured.Unstructured, finalizer string) (updated bool, err error) {
	finalizers := sets.New[string](obj.GetFinalizers()...)
	if !finalizers.Has(deletionFinalizer) {
		return false, nil
	}

	original := obj.DeepCopy()

	finalizers.Delete(deletionFinalizer)
	obj.SetFinalizers(sets.List(finalizers))

	log.Debugw("Removing finalizer…", "on", newObjectKey(obj, ""), "finalizer", finalizer)
	if err := client.Patch(ctx, obj, ctrlruntimeclient.MergeFrom(original)); err != nil {
		return false, err
	}

	return true, nil
}

type objectKey struct {
	Cluster   string
	Namespace string
	Name      string
}

func newObjectKey(obj metav1.Object, clusterName string) objectKey {
	return objectKey{
		Cluster:   clusterName,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func (k objectKey) String() string {
	result := k.Name
	if k.Namespace != "" {
		result = k.Namespace + "/" + result
	}
	if k.Cluster != "" {
		result = k.Cluster + "|" + result
	}

	return result
}

func (k objectKey) Key() string {
	result := k.Name
	if k.Namespace != "" {
		result = k.Namespace + "_" + result
	}
	if k.Cluster != "" {
		result = k.Cluster + "_" + result
	}

	return result
}

func (k objectKey) Labels() labels.Set {
	return labels.Set{
		remoteObjectClusterLabel:   k.Cluster,
		remoteObjectNamespaceLabel: k.Namespace,
		remoteObjectNameLabel:      k.Name,
	}
}
