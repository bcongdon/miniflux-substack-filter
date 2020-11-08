// The entry point for miniflux-substack-filter.
package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/bcongdon/miniflux-substack-filter/filter"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/peterbourgon/ff"
	"github.com/robfig/cron/v3"

	miniflux "miniflux.app/client"
)

func setupLogger(logLevel string) log.Logger {
	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	switch strings.ToLower(logLevel) {
	case "debug":
		l = level.NewFilter(l, level.AllowDebug())
	case "info":
		l = level.NewFilter(l, level.AllowInfo())
	case "warn":
		l = level.NewFilter(l, level.AllowWarn())
	case "error":
		l = level.NewFilter(l, level.AllowError())
	default:
		l = level.NewFilter(l, level.AllowInfo())
	}
	return log.With(l, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}

func main() {
	fs := flag.NewFlagSet("mf", flag.ExitOnError)
	var (
		minifluxUsername    = fs.String("username", "", "the username used to log into miniflux")
		minifluxPassword    = fs.String("password", "", "the password used to log into miniflux")
		minifluxAPIKey      = fs.String("api-key", "", "api key used for authentication")
		minifluxAPIEndpoint = fs.String("api-endpoint", "https://rss.notmyhostna.me", "the api of your miniflux instance")
		refreshInterval     = fs.String("refresh-interval", "", "interval defining how often we check for new entries in miniflux")
		dryRun              = fs.Bool("dry-run", false, "whether to start in dry run mode")
		logLevel            = fs.String("log-level", "", "the level to filter logs at eg. debug, info, warn, error")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("MF"),
	)

	l := setupLogger(*logLevel)

	var client *miniflux.Client
	if *minifluxUsername != "" && *minifluxPassword != "" {
		client = miniflux.New(*minifluxAPIEndpoint, *minifluxUsername, *minifluxPassword)
	} else if *minifluxAPIKey != "" {
		client = miniflux.New(*minifluxAPIEndpoint, *minifluxAPIKey)
	} else {
		level.Error(l).Log("err", errors.New("api endpoint, username and password or api key need to be provided"))
		return
	}
	u, err := client.Me()
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
	level.Info(l).Log("msg", "logged in successfully", "user_id", u.ID)

	c := cron.New()
	if *refreshInterval == "" {
		*refreshInterval = "*/5 * * * *"
		level.Info(l).Log("msg", "set fallback interval as non provided", "interval_cron", *refreshInterval)
	}

	svc, err := filter.New(client, l, *dryRun)
	if err != nil {
		level.Error(l).Log("msg", "unable to create filter service", "err", err)
		return
	}
	c.AddFunc(*refreshInterval, func() {
		level.Info(l).Log("msg", "running filter job")
		if err := svc.RunFilterJob(); err != nil {
			level.Error(l).Log("msg", "filter cron job failed", "err", err)
		}
	})
	c.Run()
}
