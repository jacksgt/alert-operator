package main

import (
	"fmt"
	"net/http"
)

// https://prometheus.io/docs/prometheus/latest/querying/api/#alerts
func prometheusAlerts(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, `{
    "data": {
        "alerts": [
            {
                "activeAt": "2018-07-04T20:27:12.60602144+02:00",
                "annotations": {},
                "labels": {
                    "alertname": "my-alert"
                },
                "state": "firing",
                "value": "1e+00"
            }
        ]
    },
    "status": "success"
}`)
}

func main() {
	http.HandleFunc("/api/v1/alerts", prometheusAlerts)

	fmt.Println("Running on port 9090 ...")
	http.ListenAndServe("localhost:9090", nil)
}
