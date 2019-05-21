package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pseidemann/finish"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lwolf/k8s-demo-app/metrics"
)

var (
	Version   string
	BuildTime string

	listenAddr   = flag.String("listen.addr", "0.0.0.0:8000", "Address to serve http from")
	metricsAddr  = flag.String("metrics.addr", "0.0.0.0:8001", "Address to serve metrics from")
	logsActivity = flag.Bool("logs.activity", false, "Set to produce fake log activity")
	prettyLogs   = flag.Bool("logs.pretty", true, "")

	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const (
	metricsURL   = "/metrics"
	livenessURL  = "/live"
	readynessURL = "/ready"
)

func ProduceLogActivity() {
	ticker := time.NewTicker(900 * time.Millisecond)
	for {
		for t := range ticker.C {
			log.Info().Msgf("Data chunk %d has been processed", t.Nanosecond())
		}
	}

}
func main() {
	flag.Parse()
	var output io.Writer
	if *prettyLogs {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	} else {
		output = os.Stdout
	}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Info().Str("version", Version).Str("buildTime", BuildTime).Msg("initializing the application")
	if *listenAddr == "" {
		log.Fatal().Msg("listen address couldn't be empty")
	}
	m := metrics.NewMetrics()

	mux := http.NewServeMux()
	handler := requestsHandler(m)
	mux.HandleFunc(livenessURL, handleLiveness)
	mux.HandleFunc(readynessURL, handleReadiness)
	mux.HandleFunc("/", handler)

	srv := &http.Server{Addr: *listenAddr, Handler: mux}
	fin := &finish.Finisher{
		Timeout: 10 * time.Second,
	}

	fin.Add(srv)

	if *metricsAddr != "" {
		mmux := http.NewServeMux()
		mmux.HandleFunc(metricsURL, promhttp.Handler().ServeHTTP)
		log.Info().Str("address", *metricsAddr).Msg("metrics server initialized")
		mSrv := &http.Server{Addr: *metricsAddr, Handler: mmux}
		fin.Add(mSrv)
		go func() {
			err := mSrv.ListenAndServe()
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("metrics server stopped")
			}
		}()

	}
	go func() {
		log.Info().Str("address", *listenAddr).Msg("web server initialized")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("web server stopped")
		}
	}()
	if *logsActivity {
		go ProduceLogActivity()
	}
	fin.Wait()

}

func getIpAddresses() (addresses []string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			addresses = append(addresses, ip.String())
		}
	}
	return
}

type Payload struct {
	Hostname string   `json:"hostname"`
	Ips      []string `json:"ips"`
	Version  string   `json:"version"`
}

func requestsHandler(m *metrics.Metrics) http.HandlerFunc {
	log.Info().Msg("inside the request handler")
	hostname, err := os.Hostname()
	if err != nil {
		log.Error().Msg("unable to get Hostname, generating something...")
		hostname = fmt.Sprintf("gen-host-%s", Version)
	}
	payload := Payload{
		Hostname: hostname,
		Ips:      getIpAddresses(),
		Version:  Version,
	}
	_, err = json.Marshal(payload)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to serialize response")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("path", r.URL.Path).Msg("processing request")
		m.RequestsTotal.Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
	}
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	delay := rng.Int63n(60) + 30
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Fprintln(w, "OK")
}

func handleLiveness(w http.ResponseWriter, r *http.Request) {
	delay := rng.Int63n(30)
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Fprintln(w, "OK")
}
