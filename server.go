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
    url, _ := url.Parse(r.Header.Get("x-iw-fwd"))
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
}

func main() {
    http.HandleFunc("/", handler)

    port := fmt.Sprintf(":%s", viper.GetString("port"))
    log.Fatal().Err(http.ListenAndServe(port, nil))
}
