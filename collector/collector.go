package collector

import (
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Collector struct {
	Router *HitronRouter
}

const prefix = "hitron_"

var (
	loginSuccessDesc *prometheus.Desc = prometheus.NewDesc(
		prefix+"login_success_bool", "1 if the login was successful", nil, nil)

	// info
	systemUpdateDesc = prometheus.NewDesc(
		prefix+"info_uptime", "System uptime", nil, nil)
)

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- loginSuccessDesc

	ch <- systemUpdateDesc
}
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	err := c.Router.Login()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(loginSuccessDesc, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(loginSuccessDesc, prometheus.GaugeValue, 1)

	info, err := c.Router.Info()
	if err != nil {
		// todo count errors
		log.Info("Info: ", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(systemUpdateDesc, prometheus.CounterValue, parseUptime(info.SystemUptime))
}

var uptimeRegex = regexp.MustCompile(`^(\d+) Days,(\d+) Hours,(\d+) Minutes,(\d+) Seconds$`)

func parseUptime(raw string) float64 {
	matches := uptimeRegex.FindStringSubmatch(raw)
	if len(matches) < 5 {
		return 0
	}
	daysR, hoursR, minutesR, secondsR := matches[1], matches[2], matches[3], matches[4]

	days, err := strconv.ParseInt(daysR, 10, 64)
	if err != nil {
		return 0
	}
	hours, err := strconv.ParseInt(hoursR, 10, 64)
	if err != nil {
		return 0
	}
	minutes, err := strconv.ParseInt(minutesR, 10, 64)
	if err != nil {
		return 0
	}
	seconds, err := strconv.ParseInt(secondsR, 10, 64)
	if err != nil {
		return 0
	}
	return float64(seconds +
		60*minutes +
		60*60*hours +
		60*60*24*days)
}
