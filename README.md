# datasource_exporter

Exporter example that checks if the datasource `appstudio-prometheus-ds` exists in our 
Grafana definition.
It returns a metric with value of 1 if it exists and 0 otherwise.

The go code uses an in-cluster client configuration to interact with the cluster.

## Prerequisites
An Openshift cluster with RHTAP deployed in (we are checking RHTAP resource)

## Build
- Clone the repository
- Build a docker image
- Push the image to registry
- Update the `image` entry in the attached yaml file
- Apply the yaml to the cluster

## Test
- Port-forward the pod (port 9101)
- Open a browser and insert the url http://localhost:9101/metrics
- Each refresh of the page triggers the exporter and will update the metric 