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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	ratelimitv1alpha1 "github.com/mingli103/k8s-controller/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var ratelimitedconsumerlog = logf.Log.WithName("ratelimitedconsumer-resource")

// SetupRateLimitedConsumerWebhookWithManager registers the webhook for RateLimitedConsumer in the manager.
func SetupRateLimitedConsumerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&ratelimitv1alpha1.RateLimitedConsumer{}).
		WithValidator(&RateLimitedConsumerCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-ratelimit-itbl-sre-co-v1alpha1-ratelimitedconsumer,mutating=false,failurePolicy=fail,sideEffects=None,groups=ratelimit.itbl.sre.co,resources=ratelimitedconsumers,verbs=create;update,versions=v1alpha1,name=vratelimitedconsumer-v1alpha1.kb.io,admissionReviewVersions=v1

// RateLimitedConsumerCustomValidator struct is responsible for validating the RateLimitedConsumer resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type RateLimitedConsumerCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &RateLimitedConsumerCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type RateLimitedConsumer.
func (v *RateLimitedConsumerCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ratelimitedconsumer, ok := obj.(*ratelimitv1alpha1.RateLimitedConsumer)
	if !ok {
		return nil, fmt.Errorf("expected a RateLimitedConsumer object but got %T", obj)
	}
	ratelimitedconsumerlog.Info("Validation for RateLimitedConsumer upon creation", "name", ratelimitedconsumer.GetName())

	return nil, validateRateLimitConsumer(ratelimitedconsumer)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type RateLimitedConsumer.
func (v *RateLimitedConsumerCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ratelimitedconsumer, ok := newObj.(*ratelimitv1alpha1.RateLimitedConsumer)
	if !ok {
		return nil, fmt.Errorf("expected a RateLimitedConsumer object for the newObj but got %T", newObj)
	}
	ratelimitedconsumerlog.Info("Validation for RateLimitedConsumer upon update", "name", ratelimitedconsumer.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, validateRateLimitConsumer(ratelimitedconsumer)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type RateLimitedConsumer.
func (v *RateLimitedConsumerCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ratelimitedconsumer, ok := obj.(*ratelimitv1alpha1.RateLimitedConsumer)
	if !ok {
		return nil, fmt.Errorf("expected a RateLimitedConsumer object but got %T", obj)
	}
	ratelimitedconsumerlog.Info("Validation for RateLimitedConsumer upon deletion", "name", ratelimitedconsumer.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

func validateRateLimitConsumer(ratelimitedconsumer *ratelimitv1alpha1.RateLimitedConsumer) error {
	var allErrs field.ErrorList

	if len(ratelimitedconsumer.Spec.RateLimit.Names) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("rateLimit").Child("names"), ratelimitedconsumer.Spec.RateLimit.Names, "rateLimit.names must not be empty"))
	}

	if ratelimitedconsumer.Spec.TargetRoute.Name == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("targetRoute").Child("name"), ratelimitedconsumer.Spec.TargetRoute.Name, "targetRoute.name must not be empty"))
	}

	if len(allErrs) > 0 {
		return errors.NewInvalid(schema.GroupKind{
			Group: "ratelimit.itbl.sre.co", Kind: "RateLimitedConsumer"},
			ratelimitedconsumer.Name, allErrs)
	}
	return nil
}
