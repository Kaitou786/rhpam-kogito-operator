// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/operator"
	rhpamv1 "github.com/kiegroup/rhpam-kogito-operator/api/v1"
	"github.com/kiegroup/rhpam-kogito-operator/internal"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KogitoRuntimeReconciler reconciles a KogitoRuntime object
type KogitoRuntimeReconciler struct {
	*client.Client
	Log    logger.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoruntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoruntimes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoruntimes/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;list;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams;imagestreamtags,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;roles,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=core,resources=configmaps;events;pods;secrets;serviceaccounts;services,verbs=create;delete;get;list;patch;update;watch

// Reconcile reads that state of the cluster for a KogitoRuntime object and makes changes based on the state read
// and what is in the KogitoRuntime.Spec
func (r *KogitoRuntimeReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling for KogitoRuntime")

	// create context
	context := &operator.Context{
		Client: r.Client,
		Log:    log,
		Scheme: r.Scheme,
	}

	// fetch the requested instance
	runtimeHandler := internal.NewKogitoRuntimeHandler(context)
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(req.NamespacedName)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debug("KogitoRuntime instance not found")
		return
	}

	rbacHandler := infrastructure.NewRBACHandler(context)
	if err = rbacHandler.SetupRBAC(req.Namespace); err != nil {
		return
	}

	healthCheckProbeType := kogitoservice.TCPHealthCheckProbe
	if instance.GetSpec().GetRuntime() == api.QuarkusRuntimeType {
		healthCheckProbeType = kogitoservice.QuarkusHealthCheckProbe
	}
	deploymentHandler := NewRuntimeDeployerHandler(context, instance, runtimeHandler)
	definition := kogitoservice.ServiceDefinition{
		Request:            req,
		DefaultImageTag:    infrastructure.LatestTag,
		SingleReplica:      false,
		OnDeploymentCreate: deploymentHandler.OnDeploymentCreate,
		OnGetComparators:   deploymentHandler.OnGetComparators,
		CustomService:      true,
		HealthCheckProbe:   healthCheckProbeType,
	}
	requeueAfter, err := kogitoservice.NewServiceDeployer(context, definition, instance, nil).Deploy()
	if err != nil {
		return
	}
	if requeueAfter > 0 {
		log.Info("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = requeueAfter
		result.Requeue = true
	}
	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoRuntime controller")

	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.MetaNew.GetDeletionTimestamp().IsZero()
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).
		For(&rhpamv1.KogitoRuntime{}, builder.WithPredicates(pred)).
		Owns(&corev1.Service{}).Owns(&appsv1.Deployment{}).Owns(&corev1.ConfigMap{})

	if r.IsOpenshift() {
		b.Owns(&routev1.Route{}).Owns(&imagev1.ImageStream{})
	}

	return b.Complete(r)
}
