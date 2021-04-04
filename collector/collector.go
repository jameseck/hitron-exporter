package collector

import (
	"fmt"
	"math"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Collector struct {
	Router *HitronRouter
}

const prefix = "hitron_"

var (
	loginSuccessDesc *prom.Desc = prom.NewDesc(
		prefix+"login_success_bool", "1 if the login was successful", nil, nil)
	scrapeTimeDesc *prom.Desc = prom.NewDesc(
		prefix+"scrape_time", "Time the scrape run took", nil, nil)

	// SysInfo
	systemUptimeDesc = prom.NewDesc(
		prefix+"info_uptime", "System uptime", nil, nil)
	versionDesc = prom.NewDesc(
		prefix+"version", "Versions in labels", []string{"hw_version", "sw_version", "serial"}, nil)
	addressDesc = prom.NewDesc(
		prefix+"address", "Hardware and IP Addresses in labels", []string{"wan_ip", "lan_ip", "rf_mac"}, nil)
	trafficDesc = prom.NewDesc(
		prefix+"traffic", "Basic traffic counters. if=wan/lan, dir=send/recv.", []string{"if", "dir"}, nil)

	// CMInit
	cmHwInitDesc = prom.NewDesc(
		prefix+"cm_hwinit_success", "DOCSIS Provisioning HWInit Status", nil, nil)
)

func (c *Collector) Describe(ch chan<- *prom.Desc) {
	ch <- loginSuccessDesc
	ch <- scrapeTimeDesc

	// SysInfo
	ch <- systemUptimeDesc
	ch <- versionDesc
	ch <- addressDesc
	ch <- trafficDesc

	// CMInit
	ch <- cmHwInitDesc
}
func (c *Collector) Collect(ch chan<- prom.Metric) {
	defer func() func() {
		startTime := time.Now()
		return func() {
			duration := time.Since(startTime).Seconds()
			ch <- prom.MustNewConstMetric(scrapeTimeDesc, prom.GaugeValue, duration)
		}
	}()()

	session, err := c.Router.Login()
	if err != nil {
		ch <- prom.MustNewConstMetric(loginSuccessDesc, prom.GaugeValue, 0)
		return
	}
	defer session.Logout()
	ch <- prom.MustNewConstMetric(loginSuccessDesc, prom.GaugeValue, 1)

	info, err := session.Info()
	if err != nil {
		// todo count errors
		log.Info("Info: ", err)
		return
	}
	ch <- prom.MustNewConstMetric(systemUptimeDesc, prom.CounterValue, parseUptime(info.SystemUptime))
	ch <- prom.MustNewConstMetric(versionDesc, prom.GaugeValue, 1, info.HwVersion, info.SwVersion, info.SerialNumber)
	ch <- prom.MustNewConstMetric(addressDesc, prom.GaugeValue, 1, info.WanIp, info.LanIp, info.RfMac)
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.LRecPkt), "lan", "recv")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.LSendPkt), "lan", "send")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.WRecPkt), "wan", "recv")
	ch <- prom.MustNewConstMetric(trafficDesc, prom.CounterValue, parsePkt(info.WSendPkt), "wan", "send")

	cmInit, err := session.CMInit()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}
	ch <- prom.MustNewConstMetric(cmHwInitDesc, prom.GaugeValue,
		is(StatusSuccess, cmInit.HwInit))

	_, err = session.CMDocsisWAN()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}

	_, err = session.ConnectInfo()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}

	_, err = session.UpstreamInfo()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}

	_, err = session.DownstreamInfo()
	if err != nil {
		log.Info("CMInit: ", err)
		return
	}
}

func is(expected, actual string) float64 {
	if expected == actual {
		return 1
	}
	return 0
}

// parseUptime parses a duration in the format of "05 Days,21 Hours,33 Minutes,44 Seconds"
// Example
func parseUptime(raw string) float64 {
	var days, hours, minutes, seconds time.Duration
	_, err := fmt.Sscanf(raw, "%2d Days,%2d Hours,%2d Minutes,%2d Seconds", &days, &hours, &minutes, &seconds)
	if err != nil {
		log.Warn("Unknown uptime format: ", err, " ", raw)
		return -1
	}

	return float64(seconds +
		minutes*60 +
		hours*3600 +
		days*24*3600)
}

var scaleMap = map[string]float64{
	"K": 1024,
	"M": math.Pow(1024, 2),
	"G": math.Pow(1024, 3),
	"T": math.Pow(1024, 4),
	"E": math.Pow(1024, 5),
}

func parsePkt(raw string) float64 {
	var out float64
	var scale string
	_, err := fmt.Sscanf(raw, "%f%1s Bytes", &out, &scale)
	if err != nil {
		return -1
	}
	factor, ok := scaleMap[scale]
	if !ok {
		log.Warn("Unknown scale factor in ", raw)
		return -1
	}
	return out * factor
}
