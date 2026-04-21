package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/konano/oss-auto-cert/internal/cert"
	"github.com/konano/oss-auto-cert/internal/config"
)

var (
	logLevel string
	sig      = make(chan os.Signal, 1)
	conf     = new(config.Config)
)

func init() {
	flag.StringVar(&logLevel, "log-level", "info", "日志等级")
	flag.StringVar(&conf.Path, "config", "", "配置文件路径")
	flag.Parse()

	log.SetReportCaller(true)
	if level, err := log.ParseLevel(logLevel); err != nil {
		log.Warnf("Invalid log level parameter: %s. Use default info level!", logLevel)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}

	conf.LoadOptions()
	conf.LoadOptionsFromEnv()
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	m, err := cert.NewAutoCert(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	m.ScheduleRun()

	// wait
	select {
	case <-sig:
		cancel()
		m.Stop()
		log.Infof("Exit.")
		os.Exit(0)
	}
}
