package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	grafanaResoruce := GetGrafanaResource()
	if strings.Contains(grafanaResoruce, "prometheus-appstudio-ds") {
		metricValue = 1
	} else {
		metricValue = 0
	}

	//Write latest value for the metric in the prometheus metric channel.
	m1 := prometheus.MustNewConstMetric(collector.dsLiveMetric, prometheus.GaugeValue, metricValue)
	m1 = prometheus.NewMetricWithTimestamp(time.Now().Add(-time.Hour), m1)
	ch <- m1
}

// get the grafna resource
func GetGrafanaResource() string {
	clientset := ClusterConfig()

	data, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/grafana.integreatly.org/v1beta1").
		Namespace("appstudio-grafana").
		Resource("grafanas").
		Name("grafana-oauth").
		DoRaw(context.TODO())
	grafanaResoruce := string(data)
	fmt.Printf(grafanaResoruce)
	if err != nil {
		fmt.Printf("Error getting resource: %v\n", err)
		os.Exit(1)
	}
	return grafanaResoruce
}

func ClusterConfig() *kubernetes.Clientset {
	//creates in cluster config
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("Error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)

	if err != nil {
		fmt.Printf("error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}
	return clientset
}

func main() {
	reg := prometheus.NewPedanticRegistry()
	dsc := newDataSourceCollector()
	reg.MustRegister(dsc)

	http.Handle("/datasource-exporter/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          reg,
		},
	))
	fmt.Printf("Start datasource exporter")
	log.Fatal(http.ListenAndServe(":9101", nil))

}
