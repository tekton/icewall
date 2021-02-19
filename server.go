package main

import (
    "fmt"
    "net/http"
    "net/url"
    "net/http/httputil"

    //
    "github.com/spf13/viper"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func handler(w http.ResponseWriter, r *http.Request) {
    host := r.Header.Get("x-iw-fwd")
    if host == "" {
        host = viper.GetString("default_host")
    }

    // header rules
    // path rules
    // query string rules

    // headers to look for as an array?

    url, _ := url.Parse(host)
    proxy := httputil.NewSingleHostReverseProxy(url)
    r.URL.Host = url.Host
    r.URL.Scheme = url.Scheme
    r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
    r.Host = url.Host
    proxy.ServeHTTP(w, r)
}

func init() {
    viper.AddConfigPath(".")
    viper.AddConfigPath("/opt/icewall/config/")
    viper.AddConfigPath("/etc/icewall/config/")
    viper.SetConfigName("icewall")

    viper.SetDefault("log_level", "info")
    viper.SetDefault("default_host", "http://localhost")

    viper_err := viper.ReadInConfig()   // Find config, read config, or else...
    if viper_err != nil {
        panic(fmt.Errorf("Fatal error config file: %s \n", viper_err))
    } else {
        fmt.Println(viper.AllKeys())
    }

    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    if viper.GetString("log_level") == "debug" {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    }

    // read basic rules from files on disk? badger?
    // subscribe to "ticker" for new rules
}

func main() {
    http.HandleFunc("/", handler)
    // http.HandleFunc("/__iw__/api/add_rule", handler)

    port := fmt.Sprintf(":%s", viper.GetString("port"))
    log.Fatal().Err(http.ListenAndServe(port, nil))
}
