package netgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	AuthURL      string `json:"authUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURI  string `json:"redirectUri"`
	Scope        string `json:"scope"`

	LogHost      string `json:"logHost"`
	LogPort      int    `json:"logPort"`
	LogProgram   string `json:"logProgram"`
	LogService   string `json:"logService"`
	LogRetention string `json:"logRetention"`
}

func CreateConfig() *Config {
	return &Config{
		AuthURL:      "",
		ClientID:     "",
		ClientSecret: "",
		RedirectURI:  "",
		Scope:        "",

		LogHost:      "",
		LogPort:      514,
		LogProgram:   "",
		LogService:   "",
		LogRetention: "",
	}
}

type AuthPlugin struct {
	next    http.Handler
	name    string
	config  *Config
	udp     *net.UDPConn
	udpAddr *net.UDPAddr
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	var conn *net.UDPConn
	var addr *net.UDPAddr
	var err error

	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.LogHost, config.LogPort))
	if err != nil {
		fmt.Println("[Plugin] UDP address resolve error:", err)
		return nil, err
	}
	conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("[Plugin] UDP dial error:", err)
		return nil, err
	}

	return &AuthPlugin{
		next:    next,
		name:    name,
		config:  config,
		udp:     conn,
		udpAddr: addr,
	}, nil
}

func (a *AuthPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	code := req.Header.Get("X-Api-Key")
	if code == "" {
		a.log("Missing 'code' in request", "ERROR", req)
		http.Error(rw, "Unauthorized: missing code", http.StatusUnauthorized)
		return
	}

	clientIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		clientIP = req.RemoteAddr
	}

	form := url.Values{}
	form.Add("client_id", a.config.ClientID)
	form.Add("client_secret", a.config.ClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", a.config.RedirectURI)
	form.Add("scope", a.config.Scope)
	form.Add("ipaddress", clientIP)

	authReq, err := http.NewRequest("POST", a.config.AuthURL, strings.NewReader(form.Encode()))
	if err != nil {
		a.log("Failed to create auth request: "+err.Error(), "ERROR", req)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	authReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(authReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		a.log(fmt.Sprintf("Authentication failed: status=%d", resp.StatusCode), "WARN", req)
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	a.log("Authentication success", "INFO", req)
	a.next.ServeHTTP(rw, req)
}

func (a *AuthPlugin) log(message, level string, req *http.Request) {
	if a.udp == nil {
		return
	}

	hostname, _ := os.Hostname()
	sessionID := req.Header.Get("X-Session-Id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%d", os.Getpid())
	}

	logData := map[string]interface{}{
		"@timestamp":   time.Now().Format(time.RFC3339),
		"source":       hostname,
		"session_name": "logger",
		"retention":    a.config.LogRetention,
		"program":      a.config.LogProgram,
		"service":      a.config.LogService,
		"level":        "LEVEL_" + level,
		"message":      fmt.Sprintf("[traefik-net-gateway] %s", message),
		"session_id":   sessionID,
	}

	payload, err := json.Marshal(logData)
	if err != nil {
		fmt.Println("[Plugin] Log marshal error:", err)
		return
	}

	_, err = a.udp.Write(payload)
	if err != nil {
		fmt.Println("[Plugin] Log send error:", err)
	}
}
