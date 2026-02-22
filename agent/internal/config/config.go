package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	ServerURL string
	AuthToken string
	NodeID    string
	Interval  time.Duration
	Once      bool
	PrintOnly bool
	LHMURL    string
}

func Load() Config {
	server := env("SERVER_URL", "https://localhost:8443")
	token := env("AUTH_TOKEN", "dev-token")
	node := env("NODE_ID", host())
	intervalStr := env("INTERVAL", "5s")
	interval := mustDuration(intervalStr)
	once := false
	printOnly := false
	lhmURL := env("LHM_URL", "http://localhost:8085/data.json")

	flag.StringVar(&server, "server", server, "server base URL")
	flag.StringVar(&token, "token", token, "auth token")
	flag.StringVar(&node, "node", node, "node id")
	flag.DurationVar(&interval, "interval", interval, "collection interval")
	flag.BoolVar(&once, "once", once, "collect once and exit")
	flag.BoolVar(&printOnly, "print-only", printOnly, "print payload and do not send")
	flag.StringVar(&lhmURL, "lhm-url", lhmURL, "LibreHardwareMonitor data.json URL")
	flag.Parse()

	return Config{
		ServerURL: server,
		AuthToken: token,
		NodeID:    node,
		Interval:  interval,
		Once:      once,
		PrintOnly: printOnly,
		LHMURL:    lhmURL,
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func host() string {
	name, _ := os.Hostname()
	if name == "" {
		return "node"
	}
	return name
}

func mustDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 5 * time.Second
	}
	return d
}