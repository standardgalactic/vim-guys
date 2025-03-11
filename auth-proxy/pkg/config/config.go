package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

const PROXY_SERVER_ID = 1

type ProxyConfigParams struct {
	DBUrl   string
	DBToken string

	Port                  uint16
	AuthenticationTimeout time.Duration
	ReadLimit             int
}

type ProxyConfigJSON struct {
	Port                    uint16 `json:"port"`
	AuthenticationTimeoutMS int    `json:"authentication_timeout_ms"`
	WSReadLimit             int    `json:"ws_read_limit"`
}

func defaultProxyParams() *ProxyConfigParams {
	return &ProxyConfigParams{
		DBUrl:   "",
		DBToken: "",

		Port:                  42000,
		AuthenticationTimeout: time.Second * 3,
		ReadLimit:             50,
	}
}

func configureFromConfigFile(config *ProxyConfigParams) {
	args := os.Args
	if len(args) < 2 {
		return
	}
	contents, err := os.Open(args[1])
	if err != nil {
		return
	}

	var params ProxyConfigJSON
	err = json.NewDecoder(contents).Decode(&params)
	if err != nil {
		return
	}

	if params.Port > 0 {
		config.Port = uint16(params.Port)
	}

	if params.AuthenticationTimeoutMS > 0 {
		config.AuthenticationTimeout = time.Millisecond * time.Duration(params.AuthenticationTimeoutMS)
	}

	if params.WSReadLimit > 0 {
		config.ReadLimit = params.WSReadLimit
	}
}

func getEnvNumber(prop string) int {
	value := os.Getenv(prop)
	if value == "" {
		return -1
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return v
}

func configureFromCLIParams(config *ProxyConfigParams) {
	port := uint(0)
	authenticationTimeout := int64(0)
	readLimit := 0

	flag.UintVar(&port, "port", 0, "the port to launch the proxy server on")
	flag.Int64Var(&authenticationTimeout, "auth-timeout", -1, "time to wait before kicking the websocket, specify in MS")
	flag.IntVar(&readLimit, "read-limit", -1, "the maximum amount of bytes receviable by client ws. default=50")
	flag.Parse()

	if port != 0 {
		config.Port = uint16(port)
	}

	if authenticationTimeout != -1 {
		config.AuthenticationTimeout = time.Millisecond * time.Duration(authenticationTimeout)
	}

	if readLimit != -1 {
		config.ReadLimit = readLimit
	}
}

func configureFromEnv(config *ProxyConfigParams) {
	port := getEnvNumber("PORT")
	authenticationTimeout := getEnvNumber("AUTHENTICATION_TIMEOUT_MS")
	readLimit := getEnvNumber("WS_READ_LIMIT")

	if port > 0 {
		config.Port = uint16(port)
	}

	if authenticationTimeout > 0 {
		config.AuthenticationTimeout = time.Millisecond * time.Duration(authenticationTimeout)
	}

	if readLimit > 0 {
		config.ReadLimit = readLimit
	}

	dbUrl := os.Getenv("TURSO_DATABASE_URL")
	dbToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbUrl != "" {
		config.DBUrl = dbUrl
	}
	if dbToken != "" {
		config.DBToken = dbToken
	}
}

func ProxyConfigParamsFromEnv() *ProxyConfigParams {
	config := defaultProxyParams()
	configureFromEnv(config)
	configureFromConfigFile(config)
	configureFromCLIParams(config)
	return config
}

type ProxyContextWS struct {
	ReadLimit             int
	AuthenticationTimeout time.Duration
}

type ProxyContext struct {
	WS     ProxyContextWS
	Port   uint16
	DB     *sqlx.DB
	Logger *slog.Logger
}

func (p *ProxyContext) HasDatabase() bool {
	return p.DB != nil
}

func (p *ProxyContext) addDB(config *ProxyConfigParams) *ProxyContext {
	connStr := fmt.Sprintf("libsql://%s?authToken=%s", config.DBUrl, config.DBToken)
	db, err := sqlx.Connect("libsql", connStr)
	if err != nil {
		p.Logger.Error("Failed to connect to Turso database", "error", err)
		return p
	}
	// Test the connection
	if err := db.Ping(); err != nil {
		p.Logger.Error("Failed to ping database", "error", err)
		db.Close()
	}

	p.Logger.Warn("Successfully connected to Turso database!")
	p.DB = db
	return p
}

func (p *ProxyContext) setWSConfig(config *ProxyConfigParams) *ProxyContext {
	p.WS = ProxyContextWS{
		ReadLimit: config.ReadLimit,
		AuthenticationTimeout: config.AuthenticationTimeout,
	}
	return p
}

func (p *ProxyContext) setTopLevelInformation(config *ProxyConfigParams) *ProxyContext {
	p.Port = config.Port

	// probably need a way to configure the slogger
	p.Logger = slog.Default()

	return p
}

func (p *ProxyContext) Close() {
	if p.DB != nil {
		err := p.DB.Close()
		if err != nil {
			p.Logger.Error("closing db resulted in error", "error", err)
		}
	}
}

func NewAuthConfig(config *ProxyConfigParams) *ProxyContext {
	ctx := &ProxyContext{}

	return ctx.
		addDB(config).
	    setWSConfig(config).
	    setTopLevelInformation(config)
}
