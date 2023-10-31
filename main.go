package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Define a struct for the collector that contains pointers
// to prometheus descriptors for each metric we want to expose
type dataSourceCollector struct {
	dsLiveMetric *prometheus.Desc
}

// Constructor for the collector that
// initializes the descriptor and returns a pointer to the collector
func newDataSourceCollector() *dataSourceCollector {
	return &dataSourceCollector{
		dsLiveMetric: prometheus.NewDesc("ds_live_metric",
			"Shows whether ds is a live",
			nil, nil,
		),
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *dataSourceCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector

	ch <- collector.dsLiveMetric
}

// Collect implements required collect function for all promehteus collectors
func (collector *dataSourceCollector) Collect(ch chan<- prometheus.Metric) {

	var metricValue float64

	metricValue = IsDataSourceExist(GetDataSources(GetGrafanaResource()), "prometheus-appstudio-ds")

	//Write latest value for the metric in the prometheus metric channel.
	m1 := prometheus.MustNewConstMetric(collector.dsLiveMetric, prometheus.GaugeValue, metricValue)
	m1 = prometheus.NewMetricWithTimestamp(time.Now().Add(-time.Hour), m1)
	ch <- m1
}

// get the grafna resource as a map
func GetGrafanaResource() map[string]interface{} {
	clientset := NewKubeClient()

	data, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/grafana.integreatly.org/v1beta1").
		Namespace("appstudio-grafana").
		Resource("grafanas").
		Name("grafana-oauth").
		DoRaw(context.TODO())
	var grafanaResoruce map[string]interface{}
	err = json.Unmarshal(data, &grafanaResoruce)
	if err != nil {
		fmt.Printf("Error getting resource: %v\n", err)
		os.Exit(1)
	}

	return grafanaResoruce
}

// get datasources from grafana resource
func GetDataSources(grafanaResource map[string]interface{}) []string {
	// return empty string slice if datasources are not defined
	if grafanaResource["status"].(map[string]any)["datasources"] == nil {
		return make([]string, 0)
	}
	datasourcesIfc := grafanaResource["status"].(map[string]any)["datasources"].([]interface{})
	datasources := make([]string, len(datasourcesIfc))
	for i, v := range datasourcesIfc {
		datasources[i] = v.(string)
	}
	return datasources
}

// check if datasource exists, return 1 if yes, 0 if not
func IsDataSourceExist(datasources []string, dsToCheck string) float64 {
	for _, datasource := range datasources {
		if strings.Contains(datasource, dsToCheck) {
			fmt.Println("Datasource", datasource, "exists")
			return 1
		}
	}
	fmt.Println("Datasource", dsToCheck, "does not exist")
	return 0
}

func NewKubeClient() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func main() {

	fmt.Println("Start datasource exporter")
	reg := prometheus.NewPedanticRegistry()
	dsc := newDataSourceCollector()
	reg.MustRegister(dsc)

	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          reg,
		},
	))
	log.Fatal(http.ListenAndServe(":9101", nil))

}
