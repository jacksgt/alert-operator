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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
)

// AlertReconciler reconciles a Alert object
type AlertReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	ControllerNamespace string
	PrometheusBaseURL   string
	SyncChannel         chan event.GenericEvent
}

// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=alerts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=alerts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=alertmanager.prometheus.io.alertmanager.prometheus.io,resources=alerts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Alert object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *AlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if req.NamespacedName.Name != "" {
		// TODO: reconcile individual alert
		return ctrl.Result{}, nil
	}

	log.Info("syncing all alerts")

	alerts, err := getActiveAlerts(ctx, r.PrometheusBaseURL)
	if err != nil {
		// error talking to prometheus, retry later
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Got %d alerts from Prometheus", len(alerts)))

	for _, a := range alerts {
		var alertObj = alertmanagerprometheusiov1alpha1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateAlertName(a),
				Namespace: r.ControllerNamespace,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &alertObj, func() error {
			return nil
		})
		if err != nil {
			log.Error(err, "Unable to create Alert", alertObj.ObjectMeta.Name)
			continue
		}

		// TODO: consider using something like CreateOrUpdate
		// Set status
		alertObj.Status.State = a.State
		alertObj.Status.Annotations = a.Annotations
		alertObj.Status.Labels = a.Labels
		alertObj.Status.Since = a.ActiveAt.String()
		alertObj.Status.Value = a.Value

		err = r.updateAlertStatus(&alertObj)
		if err != nil {
			log.Error(err, "Unable to set alert status")
			continue
		}
	}

	// TODO: garbage collect old alerts

	return ctrl.Result{}, nil
}

func (r *AlertReconciler) updateAlertStatus(a *alertmanagerprometheusiov1alpha1.Alert) error {
	if err := r.Status().Update(context.TODO(), a); err != nil {
		return fmt.Errorf("Failed to update Alert.status with err: %w", err)
	}
	return nil
}

func generateAlertName(a Alert) string {
	alertName := a.Labels["alertname"]
	if alertName == "" {
		// According to https://github.com/prometheus/prometheus/blob/d002fad00c20eaad029d6d122bfc513b091f78ad/rules/alerting.go#L394
		// the "alertname" label should always be set		
		panic("alertname label is not set!")
	}
	// TODO: include labels / annotations in calculation as well?
	data := a.ActiveAt.String()
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%s-%x", alertName, hash[0:8])
}

func getActiveAlerts(ctx context.Context, baseURL string) ([]Alert, error) {
	var alerts []Alert
	// TODO: authentication
	resp, err := http.Get(baseURL + "/api/v1/alerts")
	if err != nil {
		return alerts, fmt.Errorf("Error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return alerts, fmt.Errorf("Error reading response body: %w", err)
	}

	// Parse the JSON response
	var alertResponse PrometheusAlertResponse
	err = json.Unmarshal(body, &alertResponse)
	if err != nil {
		return alerts, fmt.Errorf("Error parsing JSON: %w", err)
	}

	if alertResponse.Status != "success" {
		return alerts, fmt.Errorf("Unexpected response status in JSON: '%s'", alertResponse.Status)
	}

	// All good, return the alerts
	return alertResponse.Data.Alerts, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// whenever we get an event on this channel, we trigger a sync ("reconciliation") for all alerts
	return ctrl.NewControllerManagedBy(mgr).
		Named("alert_controller_syncall"). // Must be compatible with a Prometheus metric name i.e. alphanum + underscore
		WatchesRawSource(
			source.Channel(r.SyncChannel, &handler.EnqueueRequestForObject{}),
		).
		Complete(r)
}

type PrometheusAlertResponse struct {
	Data struct {
		Alerts []Alert `json:"alerts"`
	} `json:"data"`
	Status string `json:"status"`
}

type Alert struct {
	ActiveAt    time.Time         `json:"activeAt"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	State       string            `json:"state"`
	Value       string            `json:"value"`
}
