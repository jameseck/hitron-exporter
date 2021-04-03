package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	Router *HitronRouter
}

const prefix = "hitron_"

var (
	loginSuccessDesc *prometheus.Desc
)

func (c *Collector) Describe(chan<- *prometheus.Desc) {

}
func (c *Collector) Collect(chan<- prometheus.Metric) {

}
