package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cfstras/hitron-exporter/collector"
)

func main() {
	flags := pflag.NewFlagSet("server", pflag.ExitOnError)

	flags.String("host", "http://192.168.0.1:80", "Host and port to connect to router")
	flags.StringP("user", "u", "admin", "Login username")
	flags.StringP("pass", "p", "admin", "Login password")
	flags.BoolP("debug", "d", false, "Enable debug mode")
	flags.StringP("bind", "b", ":80", "HTTP Bind address for metrics")

	flags.Parse(os.Args)
	os.Args = os.Args[0:1] // clear arguments for coredns
	viper.BindPFlags(flags)

	viper.SetEnvPrefix("HIT") // will be uppercased automatically
	viper.AutomaticEnv()

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	startServer()
}

func startServer() {
	log.Infoln("Starting hitron-exporter")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>hitron-exporter</title></head>
            <body>
            <h1>hitron-exporter</h1>
            <a href="/metrics">metrics</a>
            </body>
            </html>`))
	})
	http.HandleFunc("/metrics", handleMetricsRequest)

	bindHost := viper.GetString("bind")
	log.Infoln("Listening on", bindHost)
	log.Fatal(http.ListenAndServe(bindHost, nil))
}

func handleMetricsRequest(w http.ResponseWriter, request *http.Request) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(&collector.Collector{
		Router: collector.NewHitronRouter(viper.GetString("host"), viper.GetString("user"), viper.GetString("pass")),
	})
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(),
		ErrorHandling: promhttp.ContinueOnError,
	}).ServeHTTP(w, request)
}
