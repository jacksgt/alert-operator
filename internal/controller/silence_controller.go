/*
Copyright 2024.

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

	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	alertmanagerprometheusiov1alpha1 "github.com/jacksgt/alert-operator/api/v1alpha1"
	"github.com/jacksgt/alert-operator/internal/alertmanagerapi"
)

// SilenceReconciler reconciles a Silence object
type SilenceReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Namespace          string
	SyncChannel        chan event.GenericEvent
	AlertmanagerClient *alertmanagerapi.APIClient
}

// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=silences,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=silences/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=silences/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Silence object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *SilenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(5).Info("Entering SilenceController Reconciler", "request", req)

	// sync existing silences from alertmanager server
	if req.NamespacedName.Name == "" && req.NamespacedName.Namespace == "" {
		log.V(5).Info("Running reconciliation to update all Silences from Alertmanager")
		silencesResp, httpResp, err := r.AlertmanagerClient.SilenceAPI.GetSilences(ctx).Execute()
		if err != nil {
			return ctrl.Result{}, err
		}
		_ = httpResp
		log.V(5).Info(fmt.Sprintf("Alertmanager returned %d silences", len(silencesResp)))
		for _, s := range silencesResp {
			silence := alertmanagerprometheusiov1alpha1.Silence{}
			silence.Name = s.GetId() // TODO: better name for silences?
			silence.Namespace = r.Namespace
			silence.Spec.Comment = s.GetComment()
			silence.Spec.CreatedBy = s.GetCreatedBy()
			silence.Spec.StartsAt = metav1.NewTime(s.GetStartsAt())
			silence.Spec.EndsAt = metav1.NewTime(s.GetEndsAt())
			silence.Spec.MatchLabels = map[string]string{}
			for _, matcher := range s.GetMatchers() {
				silence.Spec.MatchLabels[matcher.Name] = matcher.Value
				// TODO: handle regexes
			}

			fmt.Printf("FOO: %+v\n", silence)
			err := r.Update(ctx, &silence)
			if errors.IsNotFound(err) {
				// need to create it first
				err = r.Create(ctx, &silence)
			}
			if err != nil {
				log.Error(err, "Failed to create or update silence", "name", silence.Name, "namespace", silence.Namespace)
				return ctrl.Result{Requeue: true}, err
			}
			log.V(5).Info("Created Silence in Kubernetes after fetching it from Alertmanager", "name", silence.Name, "namespace", silence.Namespace)
		}
		// all good, exit reconciliation here
		return ctrl.Result{}, nil
	}

	// Fetch the Silence
	silence := alertmanagerprometheusiov1alpha1.Silence{}
	if err := r.Get(ctx, req.NamespacedName, &silence); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Deletions: On resource deletion, it deletes from Authzsvc API. If it doesn't exist there, it removes the finalizer
	if silence.GetDeletionTimestamp() != nil {
		// should only be deleted from API if there is a finalizer
		if !controllerutil.ContainsFinalizer(&silence, "alert-operator") {
			// nothing to do for us
			return ctrl.Result{}, nil
		}

		httpResp, err := r.AlertmanagerClient.SilenceAPI.DeleteSilence(ctx, silence.Status.SilenceId).Execute()
		_ = httpResp
		if err != nil {
			log.Error(err, "Failed to delete silence")
			// retry later
			return ctrl.Result{Requeue: true}, err
		}

		controllerutil.RemoveFinalizer(&silence, "alert-operator")
		err = r.Update(ctx, &silence)
		if err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{Requeue: true}, err
		}

		// all good, we did our job
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil

	s := generateAlertmanagerSilence(silence)
	var silenceId string

	// check if silence already exists in Alertmanager
	if silence.Status.SilenceId != "" {
		silenceResp, httpResp, err := r.AlertmanagerClient.SilenceAPI.GetSilence(ctx, silence.Status.SilenceId).Execute()
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		_ = httpResp
		// TODO: silence already exists, but let's make sure it's up-to-date
		silenceId = silenceResp.GetId()
	} else {
		// create silence in Alertmanager
		silenceResp, httpResp, err := r.AlertmanagerClient.SilenceAPI.
			PostSilences(ctx).
			Silence(convertSilenceToPost(s)).
			Execute()
		if err != nil {
 			log.Error(err, "Failed to create silence")
			return ctrl.Result{Requeue: true}, err
		}
		_ = httpResp
		silenceId = silenceResp.GetSilenceID()
	}

	// populate the object
	silence.Status.SilenceId = silenceId
	setLabel(&silence, "alertmanager.prometheus.io/silenceID", silenceId)
	controllerutil.AddFinalizer(&silence, "alert-operator")
	// TODO: set status conditions

	return ctrl.Result{}, nil
}

func convertSilenceToPost(in alertmanagerapi.Silence) alertmanagerapi.PostableSilence {
	return alertmanagerapi.PostableSilence{
		Matchers:  in.Matchers,
		StartsAt:  in.StartsAt,
		EndsAt:    in.EndsAt,
		CreatedBy: in.CreatedBy,
		Comment:   in.Comment,
	}
}

func generateAlertmanagerSilence(silence alertmanagerprometheusiov1alpha1.Silence) alertmanagerapi.Silence {
	s := alertmanagerapi.NewSilenceWithDefaults()
	s.Comment = silence.Spec.Comment
	for k, v := range silence.Spec.MatchLabels {
		m := alertmanagerapi.NewMatcher(k, v, false) // TODO: implement regex support
		s.Matchers = append(s.Matchers, *m)
	}
	s.EndsAt = silence.Spec.EndsAt.Time
	s.StartsAt = silence.Spec.StartsAt.Time

	return *s
}

// Sets a label on the resource without removing existing labels
func setLabel(obj metav1.Object, label string, value string) {
	labels := obj.GetLabels()
	labels[label] = value
	obj.SetLabels(labels)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SilenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// watch all Silence objects
		For(&alertmanagerprometheusiov1alpha1.Silence{}).
		// in addition, refresh silences from Alertmanager periodically
		WatchesRawSource(
			source.Channel(r.SyncChannel, &handler.EnqueueRequestForObject{}),
		).
		Complete(r)
}

// // https://github.com/prometheus/alertmanager/blob/main/api/v2/openapi.yaml
// func deleteSilence(id string) error {
// 	// TODO
// 	return nil
// }

// func createSilence(s AlertmanagerSilence) (string, error) {
// 	// TODO
// 	return "", nil
// }

// func getSilence(id string) (AlertmanagerSilence, error) {
// 	return AlertmanagerSilence{}, nil
// }
// func getAllSilences() ([]AlertmanagerSilence, error) {
// 	return []AlertmanagerSilence{}, nil
// }

// func compareSilence(a, b AlertmanagerSilence) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}

// 	for k, _ := range a {
// 		if a[k] != b[k] {
// 			return false
// 		}
// 	}

// 	return true
// }

// func NewSilenceController(syncInterval string, alertmanager string) *SilenceController {
// 	// https://github.com/cert-manager/openshift-routes/blob/main/internal/controller/controller.go
// 	c := SilenceReconciler{
// 		Client: mgr.GetClient(),
// 		Scheme: mgr.GetScheme(),
// 		SyncChannel         chan event.GenericEvent, // TODO:
// 		AlertmanagerClient  *alertmanagerapi.APIClient,
// 	}
// 	return &c
// }
