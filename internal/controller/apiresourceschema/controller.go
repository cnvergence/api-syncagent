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

package apiresourceschema

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/kcp-dev/logicalcluster/v3"
	"go.uber.org/zap"

	"github.com/kcp-dev/api-syncagent/internal/controllerutil/predicate"
	"github.com/kcp-dev/api-syncagent/internal/discovery"
	"github.com/kcp-dev/api-syncagent/internal/projection"
	syncagentv1alpha1 "github.com/kcp-dev/api-syncagent/sdk/apis/syncagent/v1alpha1"

	kcpdevv1alpha1 "github.com/kcp-dev/kcp/sdk/apis/apis/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/kontext"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ControllerName = "syncagent-apiresourceschema"
)

type Reconciler struct {
	localClient   ctrlruntimeclient.Client
	kcpClient     ctrlruntimeclient.Client
	log           *zap.SugaredLogger
	recorder      record.EventRecorder
	lcName        logicalcluster.Name
	agentName     string
	apiExportName string
}

// Add creates a new controller and adds it to the given manager.
func Add(
	mgr manager.Manager,
	kcpCluster cluster.Cluster,
	lcName logicalcluster.Name,
	log *zap.SugaredLogger,
	numWorkers int,
	agentName string,
	apiExportName string,
	prFilter labels.Selector,
) error {
	reconciler := &Reconciler{
		localClient:   mgr.GetClient(),
		kcpClient:     kcpCluster.GetClient(),
		lcName:        lcName,
		log:           log.Named(ControllerName),
		recorder:      mgr.GetEventRecorderFor(ControllerName),
		agentName:     agentName,
		apiExportName: apiExportName,
	}

	_, err := builder.ControllerManagedBy(mgr).
		Named(ControllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: numWorkers}).
		// Watch for changes to PublishedResources on the local service cluster
		For(&syncagentv1alpha1.PublishedResource{}, builder.WithPredicates(predicate.ByLabels(prFilter))).
		Build(reconciler)
	return err
}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.With("publishedresource", request)
	log.Debug("Processing")

	pubResource := &syncagentv1alpha1.PublishedResource{}
	if err := r.localClient.Get(ctx, request.NamespacedName, pubResource); err != nil {
		return reconcile.Result{}, ctrlruntimeclient.IgnoreNotFound(err)
	}

	// There is no special cleanup. When a PublishedResource is deleted, the
	// APIResourceSchema in kcp should remain, otherwise we risk deleting all
	// users' data just because a service admin might temporarily accidentally
	// delete the PublishedResource.
	if pubResource.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	result, err := r.reconcile(ctx, log, pubResource)
	if err != nil {
		r.recorder.Event(pubResource, corev1.EventTypeWarning, "ReconcilingError", err.Error())
	}
	if result == nil {
		result = &reconcile.Result{}
	}

	return *result, err
}

func (r *Reconciler) reconcile(ctx context.Context, log *zap.SugaredLogger, pubResource *syncagentv1alpha1.PublishedResource) (*reconcile.Result, error) {
	// find the resource that the PublishedResource is referring to
	localGVK := projection.PublishedResourceSourceGVK(pubResource)

	crd, err := discovery.NewClient(r.localClient).DiscoverResourceType(ctx, localGVK.GroupKind())
	if err != nil {
		return nil, fmt.Errorf("failed to discover resource defined in PublishedResource: %w", err)
	}

	// project the CRD
	projectedCRD, err := r.applyProjection(r.apiExportName, crd, pubResource)
	if err != nil {
		return nil, fmt.Errorf("failed to apply projection rules: %w", err)
	}

	// to prevent changing the source GVK e.g. from "apps/v1 Daemonset" to "core/v1 Pod",
	// we include the source GVK in hashed form in the final APIResourceSchema name.
	arsName := r.getAPIResourceSchemaName(r.apiExportName, projectedCRD)

	// ARS'es cannot be updated, their entire spec is immutable. For now we do not care about
	// CRDs being updated on the service cluster, but in the future (TODO) we must allow
	// service owners to somehow publish updated CRDs without changing their API version.
	wsCtx := kontext.WithCluster(ctx, r.lcName)
	ars := &kcpdevv1alpha1.APIResourceSchema{}
	err = r.kcpClient.Get(wsCtx, types.NamespacedName{Name: arsName}, ars, &ctrlruntimeclient.GetOptions{})

	if apierrors.IsNotFound(err) {
		if err := r.createAPIResourceSchema(wsCtx, log, r.apiExportName, projectedCRD, arsName); err != nil {
			return nil, fmt.Errorf("failed to create APIResourceSchema: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check for APIResourceSchema: %w", err)
	}

	// Update Status with ARS name
	if pubResource.Status.ResourceSchemaName != arsName {
		original := pubResource.DeepCopy()
		pubResource.Status.ResourceSchemaName = arsName

		if !reflect.DeepEqual(original, pubResource) {
			log.Info("Patching PublishedResource status…")
			if err := r.localClient.Status().Patch(ctx, pubResource, ctrlruntimeclient.MergeFrom(original)); err != nil {
				return nil, fmt.Errorf("failed to update PublishedResource status: %w", err)
			}
		}
	}

	return nil, nil
}

func (r *Reconciler) createAPIResourceSchema(ctx context.Context, log *zap.SugaredLogger, apigroup string, projectedCRD *apiextensionsv1.CustomResourceDefinition, arsName string) error {
	// prefix is irrelevant as the reconciling framework will use arsName anyway
	converted, err := kcpdevv1alpha1.CRDToAPIResourceSchema(projectedCRD, "irrelevant")
	if err != nil {
		return fmt.Errorf("failed to convert CRD: %w", err)
	}

	ars := &kcpdevv1alpha1.APIResourceSchema{}
	ars.Name = arsName
	ars.Annotations = map[string]string{
		syncagentv1alpha1.SourceGenerationAnnotation: fmt.Sprintf("%d", projectedCRD.Generation),
		syncagentv1alpha1.AgentNameAnnotation:        r.agentName,
	}
	ars.Labels = map[string]string{
		syncagentv1alpha1.APIGroupLabel: apigroup,
	}
	ars.Spec.Group = converted.Spec.Group
	ars.Spec.Names = converted.Spec.Names
	ars.Spec.Scope = converted.Spec.Scope
	ars.Spec.Versions = converted.Spec.Versions

	log.With("name", arsName).Info("Creating APIResourceSchema…")

	return r.kcpClient.Create(ctx, ars)
}

func (r *Reconciler) applyProjection(apiGroup string, crd *apiextensionsv1.CustomResourceDefinition, pr *syncagentv1alpha1.PublishedResource) (*apiextensionsv1.CustomResourceDefinition, error) {
	result := crd.DeepCopy()
	result.Spec.Group = apiGroup

	// At this moment we ignore every non-selected version in the CRD, as we have not fully
	// decided on how to support the API version lifecycle yet. Having multiple versions in
	// the CRD will make kcp require a `conversion` to also be configured. Since we cannot
	// enforce that and want to instead work with existing CRDs as best as we can, we chose
	// this option (instead of error'ing out if a conversion is missing).
	result.Spec.Conversion = nil
	result.Spec.Versions = slices.DeleteFunc(result.Spec.Versions, func(v apiextensionsv1.CustomResourceDefinitionVersion) bool {
		return v.Name != pr.Spec.Resource.Version
	})

	if len(result.Spec.Versions) != 1 {
		// This should never happen because of checks earlier in the reconciler.
		return nil, fmt.Errorf("invalid CRD: cannot find selected version %q", pr.Spec.Resource.Version)
	}

	result.Spec.Versions[0].Served = true
	result.Spec.Versions[0].Storage = true

	projection := pr.Spec.Projection
	if projection == nil {
		return result, nil
	}

	if projection.Version != "" {
		result.Spec.Versions[0].Name = projection.Version
	}

	if projection.Kind != "" {
		result.Spec.Names.Kind = projection.Kind
		result.Spec.Names.ListKind = projection.Kind + "List"

		result.Spec.Names.Singular = strings.ToLower(result.Spec.Names.Kind)
		result.Spec.Names.Plural = result.Spec.Names.Singular + "s"
	}

	if projection.Plural != "" {
		result.Spec.Names.Plural = projection.Plural
	}

	if projection.Scope != "" {
		result.Spec.Scope = apiextensionsv1.ResourceScope(projection.Scope)
	}

	if projection.Categories != nil {
		result.Spec.Names.Categories = projection.Categories
	}

	if projection.ShortNames != nil {
		result.Spec.Names.ShortNames = projection.ShortNames
	}

	return result, nil
}

// getAPIResourceSchemaName generates the name for the ARS in kcp. Note that
// kcp requires, just like CRDs, that ARS are named following a specific pattern.
func (r *Reconciler) getAPIResourceSchemaName(apiGroup string, crd *apiextensionsv1.CustomResourceDefinition) string {
	hash := sha1.New()
	if err := json.NewEncoder(hash).Encode(crd.Spec.Names); err != nil {
		// This is not something that should ever happen at runtime and is also not
		// something we can really gracefully handle, so crashing and restarting might
		// be a good way to signal the service owner that something is up.
		panic(fmt.Sprintf("Failed to hash PublishedResource source: %v", err))
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	// include a leading "v" to prevent SHA-1 hashes with digits to break the name
	return fmt.Sprintf("v%s.%s.%s", checksum[:8], crd.Spec.Names.Plural, apiGroup)
}
