package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	AuthToken   string
	CORSOrigins string
	TLSCertPath string
	TLSKeyPath  string
	TLSEnabled  bool
}

func LoadFromEnv() Config {
	addr := getenv("HTTP_ADDR", ":8080")
	db := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable")
	token := getenv("AUTH_TOKEN", "dev-token")
	origins := getenv("CORS_ORIGINS", "*")
	tlsCert := getenv("TLS_CERT", "")
	tlsKey := getenv("TLS_KEY", "")
	tlsEnabled := tlsCert != "" && tlsKey != ""

	return Config{
		HTTPAddr:    addr,
		DatabaseURL: db,
		AuthToken:   token,
		CORSOrigins: origins,
		TLSCertPath: tlsCert,
		TLSKeyPath:  tlsKey,
		TLSEnabled:  tlsEnabled,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
