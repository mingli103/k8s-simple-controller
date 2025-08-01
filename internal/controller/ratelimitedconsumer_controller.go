/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/mingli103/k8s-simple-controller/api/v1alpha1"
	ratelimitv1alpha1 "github.com/mingli103/k8s-simple-controller/api/v1alpha1"
)

// RateLimitedConsumerReconciler reconciles a RateLimitedConsumer object
type RateLimitedConsumerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ratelimit.test.annotation.com,resources=ratelimitedconsumers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ratelimit.test.annotation.com,resources=ratelimitedconsumers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ratelimit.test.annotation.com,resources=ratelimitedconsumers/finalizers,verbs=update
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RateLimitedConsumer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *RateLimitedConsumerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx).WithValues("name", req.Name, "namespace", req.Namespace)
	logger.Info("Reconciling RateLimitedConsumer")

	var rlc ratelimitv1alpha1.RateLimitedConsumer
	var route gatewayv1.HTTPRoute

	if err := r.Get(ctx, req.NamespacedName, &rlc); err != nil {
		// logger.Error(err, "unable to fetch RateLimitedConsumer")
		// return ctrl.Result{}, client.IgnoreNotFound(err)
		if errors.IsNotFound(err) {
			// Resource was deleted, nothing to do.
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch RateLimitedConsumer")
		return ctrl.Result{}, err
	}
	// Fetch the target HTTPRoute

	routeKey := types.NamespacedName{
		Name:      rlc.Spec.TargetRoute.Name,
		Namespace: rlc.Namespace,
	}
	if err := r.Get(ctx, routeKey, &route); err != nil {
		logger.Error(err, "failed to get HTTPRoute")
		return ctrl.Result{}, err
	}

	// Ensure annotations map exists
	if route.Annotations == nil {
		route.Annotations = map[string]string{}
	}

	// Preserve existing plugins if any
	pluginNames := rlc.Spec.RateLimit.Names
	existingPlugins := route.Annotations["konghq.com/plugins"]

	// Use a set to track plugins for deduplication
	pluginSet := map[string]struct{}{}

	// If existing annotation is non-empty, initialize the set with its contents
	if existingPlugins != "" {
		for _, p := range strings.Split(existingPlugins, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				pluginSet[trimmed] = struct{}{}
			}
			logger.Info("Plugin already present in annotations", "plugin", p)
		}
	} else {
		logger.Info("No existing plugins found in annotations")
	}

	// Add plugins from the RateLimit spec
	for _, newPlugin := range pluginNames {
		trimmed := strings.TrimSpace(newPlugin)
		if trimmed == "" {
			continue
		}
		if _, exists := pluginSet[trimmed]; exists {
			logger.Info("Plugin already exists in annotation set", "plugin", trimmed)
			continue
		}
		logger.Info("Adding plugin to annotation set", "plugin", trimmed)
		pluginSet[trimmed] = struct{}{}
	}

	// Convert back to comma-separated string
	var mergedPlugins []string
	for p := range pluginSet {
		mergedPlugins = append(mergedPlugins, p)
	}
	sort.Strings(mergedPlugins) // optional for consistency
	route.Annotations["konghq.com/plugins"] = strings.Join(mergedPlugins, ",")

	// Log final annotation value
	logger.Info("Final plugin annotation", "route", route.Name, "plugins", route.Annotations["konghq.com/plugins"])

	// Apply the update
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var latest gatewayv1.HTTPRoute
		if err := r.Get(ctx, client.ObjectKeyFromObject(&route), &latest); err != nil {
			return err
		}
		latest.Annotations = route.Annotations
		if err := r.Update(ctx, &latest); err != nil {
			logger.Error(err, "failed to update HTTPRoute")
			return err
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update the RateLimitedConsumer status

	condition := metav1.Condition{
		Type:               "PluginApplied",
		Status:             metav1.ConditionTrue,
		Reason:             "AnnotationSuccessful",
		Message:            "Successfully annotated HTTPRoute with Kong rate-limit plugin",
		LastTransitionTime: metav1.Now(),
	}

	// Update status
	rlc.Status.Conditions = []metav1.Condition{condition}
	rlc.Status.ObservedRoute = route.Name
	rlc.Status.PluginApplied = strings.Join(pluginNames, ",")
	status_err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var latest ratelimitv1alpha1.RateLimitedConsumer
		if err := r.Get(ctx, client.ObjectKeyFromObject(&rlc), &latest); err != nil {
			return err
		}
		latest.Status = rlc.Status
		if err := r.Status().Update(ctx, &latest); err != nil {
			logger.Error(err, "failed to update RateLimitedConsumer status")
			return err
		}
		logger.Info("Updated RateLimitedConsumer status with PluginApplied condition")
		return nil
	})
	if status_err != nil {
		logger.Error(status_err, "failed to update status after retry")
		return ctrl.Result{}, status_err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitedConsumerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.RateLimitedConsumer{}).
		Watches(
			&gatewayv1.HTTPRoute{},
			handler.EnqueueRequestsFromMapFunc(r.mapHTTPRouteToRLC),
			builder.WithPredicates(PluginAnnotationChanged),
		).
		Owns(&gatewayv1.HTTPRoute{}).
		Named("ratelimitedconsumer").
		Complete(r)
}

var PluginAnnotationChanged = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldVal := e.ObjectOld.GetAnnotations()["konghq.com/plugins"]
		newVal := e.ObjectNew.GetAnnotations()["konghq.com/plugins"]
		if oldVal != newVal {
			fmt.Printf("[predicate] Plugin annotation changed: old=%s, new=%s, route=%s/%s\n",
				oldVal, newVal,
				e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			return true
		}
		return false
	},
	CreateFunc: func(e event.CreateEvent) bool {
		if val, exists := e.Object.GetAnnotations()["konghq.com/plugins"]; exists {
			fmt.Printf("[predicate] Plugin annotation found on new route: %s/%s => %s\n",
				e.Object.GetNamespace(), e.Object.GetName(), val)
			return true
		}
		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return false
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

func (r *RateLimitedConsumerReconciler) mapHTTPRouteToRLC(ctx context.Context, obj client.Object) []reconcile.Request {
	fmt.Printf("mapHTTPRouteToRLC")
	route, ok := obj.(*gatewayv1.HTTPRoute)
	if !ok {
		return nil
	}

	var list v1alpha1.RateLimitedConsumerList
	if err := r.List(ctx, &list, client.InNamespace(route.Namespace)); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, rlc := range list.Items {
		if rlc.Spec.TargetRoute.Name == route.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: client.ObjectKey{
					Name:      rlc.Name,
					Namespace: rlc.Namespace,
				},
			})
		}
	}

	return requests
}
