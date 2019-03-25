package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vlamug/accesslog-exporter/cache"
	"github.com/vlamug/accesslog-exporter/config"
	"github.com/vlamug/accesslog-exporter/exporter"
	"github.com/vlamug/accesslog-exporter/exposer"
	"github.com/vlamug/accesslog-exporter/input"
	"github.com/vlamug/accesslog-exporter/parser"
	"github.com/vlamug/accesslog-exporter/pkg/logging"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultWebAddress    = ":9032"
	defaultSyslogAddress = ":9033"
)

var (
	configPath          = flag.String("config.path", "etc/config.yaml", "Exporter configuration file path.")
	regexPath           = flag.String("ua-regex.path", "etc/regexes.yaml", "User agent regexes file path.")
	webListenAddress    = flag.String("web.addr", defaultWebAddress, "Address on which to expose metrics and web interface.")
	syslogListenAddress = flag.String("syslog.addr", defaultSyslogAddress, "Listen address of syslog server.")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(logging.NewContext(context.Background()))
	defer func() {
		cancel()
	}()

	logger := logging.WithContext(ctx)

	cfg, err := config.MakeConfigFromFile(*configPath)
	if err != nil {
		logger.Sugar().Fatalf("could not make config from file: %s", err)
	}

	uaParser, err := parser.NewUAParser(*regexPath)
	if err != nil {
		logger.Sugar().Fatalf("could not initialize user user agent parser: %s", err)
	}

	cc, err := cache.NewLRUCache(cfg.Global.UserAgentCacheSize)
	if err != nil {
		logger.Sugar().Fatalf("could not initialize cache: %s", err)
	}

	// create exporter
	exp, err := exporter.NewExporter(cfg, input.NewSyslog(*syslogListenAddress), parser.ParsePipedFormat, uaParser, cc, exposer.PromExposer)
	if err != nil {
		logger.Sugar().Fatalf("could not initialize exporter: %s", err)
	}

	// run exporter
	exp.Run(ctx)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-signals
		switch sig {
		case syscall.SIGHUP:
			os.Exit(0)
		case syscall.SIGKILL:
			os.Exit(0)
		case syscall.SIGINT:
			os.Exit(0)
		default:
			// do nothing
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	logger.Sugar().Infof("Web listen address: %s", *webListenAddress)
	logger.Sugar().Infof("Syslog listen address: %s", *syslogListenAddress)
	logger.Sugar().Fatal(http.ListenAndServe(*webListenAddress, nil))
}
