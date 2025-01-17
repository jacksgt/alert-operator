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

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	alertmanagerprometheusiov1alpha1 "github.com/jacksgt/alert-operator/api/v1alpha1"
	"github.com/jacksgt/alert-operator/internal/alertmanagerapi"
	"github.com/jacksgt/alert-operator/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(alertmanagerprometheusiov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var err error
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var controllerNamespace string
	var alertmanagerBaseUrl string
	var alertmanagerBearerAuthorizationToken string
	var syncInterval string
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.StringVar(&controllerNamespace, "namespace", "", "The namespace in which the controller runs and creates objects.")
	flag.StringVar(&alertmanagerBaseUrl, "alertmanager-base-url", "http://localhost:9091", "The address at which Alertmanager listens for requests.")
	flag.StringVar(&alertmanagerBearerAuthorizationToken, "alertmanager-bearer-authorization-token", "", "Bearer Authorization for authenticating with Alertmanager (optional)")
	flag.StringVar(&syncInterval, "sync-interval", "15s", "The interval at which alerts should be loaded from the Prometheus API (as a Go duration).")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// TODO: if unset, attempt autodetection
	if controllerNamespace == "" {
		// https://github.com/operator-framework/operator-sdk/blob/v0.19.4/pkg/k8sutil/k8sutil.go#L64
		// controllerNamespace, err = k8sutil.GetOperatorNamespace()
		// if err != nil {
		setupLog.Error(err, "Failed to auto-detect namespace, please specify via command line.")
		os.Exit(1)
		// }
	}

	// syncAlertsChannel, err := setupChannelWithInterval(syncInterval)
	// if err != nil {
	// 	setupLog.Error(err, "Failed to setup sync interval")
	// 	os.Exit(1)
	// }

	syncSilencesChannel, err := setupChannelWithInterval(syncInterval)
	if err != nil {
		setupLog.Error(err, "Failed to setup sync interval")
		os.Exit(1)
	}

	// TOOD: make tlsSkipVerify configurable
	alertmanagerClient := newAlertmanagerClient(alertmanagerBaseUrl, alertmanagerBearerAuthorizationToken, true)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
		// More info:
		// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/server
		// - https://book.kubebuilder.io/reference/metrics.html
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			// TODO(user): TLSOpts is used to allow configuring the TLS config used for the server. If certificates are
			// not provided, self-signed certificates will be generated by default. This option is not recommended for
			// production environments as self-signed certificates do not offer the same level of trust and security
			// as certificates issued by a trusted Certificate Authority (CA). The primary risk is potentially allowing
			// unauthorized access to sensitive metrics data. Consider replacing with CertDir, CertName, and KeyName
			// to provide certificates, ensuring the server communicates using trusted and secure certificates.
			TLSOpts: tlsOpts,
			// FilterProvider is used to protect the metrics endpoint with authn/authz.
			// These configurations ensure that only authorized users and service accounts
			// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
			// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/filters#WithAuthenticationAndAuthorization
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "970cc10b.alertmanager.prometheus.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// if err = (&controller.AlertReconciler{
	// 	Client:              mgr.GetClient(),
	// 	Scheme:              mgr.GetScheme(),
	// 	ControllerNamespace: controllerNamespace,
	// 	PrometheusBaseURL:   prometheusBaseURL,
	// 	SyncChannel: syncAlertsChannel,
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "Alert")
	// 	os.Exit(1)
	// }

	if err = (&controller.SilenceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		Namespace:          controllerNamespace,
		SyncChannel:        syncSilencesChannel,
		AlertmanagerClient: alertmanagerClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Silence")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// Sets up a Go routine that will send period events (according to the specified) interval on the channel.
func setupChannelWithInterval(interval string) (chan event.GenericEvent, error) {
	var c chan event.GenericEvent

	// parse interval
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse '%s' as duration: %w", interval, err)
	}

	// create a channel
	c = make(chan event.GenericEvent)

	// setup ticker according to the specified interval
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			// after the interval, send an event on the channel
			case <-ticker.C:
				event := event.GenericEvent{
					Object: &corev1.Event{},
				}
				fmt.Println("Sending dummy event")
				c <- event

			}
		}
	}()

	// return the channel to the caller
	return c, nil
}

func newAlertmanagerClient(baseUrl string, bearerAuthorizationToken string, tlsSkipVerify bool) *alertmanagerapi.APIClient {
	// if necessary, disable tls certificate verification
	tr := &http.Transport{		
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkipVerify,
		},
	}
	httpClient := &http.Client{Transport: tr}
	
	cfg := alertmanagerapi.NewConfiguration()
	// TODO: leave URL alone, set cfg.{Host,Scheme} instead
	cfg.Servers[0].URL = baseUrl + "/api/v2"
	cfg.UserAgent = "alert-operator/" + cfg.UserAgent

	// if necessary, add authentication header(s)
	if bearerAuthorizationToken != "" {
		cfg.AddDefaultHeader("Authorization", "Bearer "+bearerAuthorizationToken)
	}
	cfg.HTTPClient = httpClient
	
	// TODO: test the client before returning it
	return alertmanagerapi.NewAPIClient(cfg)
}
