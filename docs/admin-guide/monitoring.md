
The monitoring stack in karina includes Prometheus, Grafana, AlertManager, Thanos, and Canary-Checker

The prom
https://coreos.com/operators/prometheus/docs/latest/api.html


### Creating custom dashboards:

    You can login to grafana using root/secret to create the dashboard, then export the JSON , then create a GrafanaDashboard CRD

https://github.com/integr8ly/grafana-operator/blob/master/deploy/examples/dashboards/SimpleDashboard.yaml
