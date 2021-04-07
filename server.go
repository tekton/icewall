package main

import (
    "fmt"
    "net/http"
    "net/url"
    "net/http/httputil"
    "encoding/json"
    "time"

    //
    "github.com/rs/xid"
    "github.com/spf13/viper"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // laod balancer health short circuit rule check
    if (r.URL.Path == viper.GetString("health_check.path")) {
        log.Debug().Msg("health check")
        // we have a health check to perform! Now, which kind?
        if (viper.GetString("health_check.action") == "file") {
            // ah, a file- we should serve it back and exit early!
            w.Header().Set("Content-Type", viper.GetString("health_check.type"))
            http.ServeFile(w, r, viper.GetString("health_check.file"))
            return
        }
    }
    //
    id := xid.New()
    r.Header.Set("x-iw-id", id.String())
    host := r.Header.Get("x-iw-fwd")
    if host == "" {
        host = viper.GetString("default_host")
    }

    // log.Debug().Str("host", host).Msg("Test")

    // header rules
    // path rules
    // query string rules

    // headers to look for as an array?

    url, _ := url.Parse(host)
    proxy := httputil.NewSingleHostReverseProxy(url)
    // defer proxy.Close()

    r.URL.Host = url.Host
    r.URL.Scheme = url.Scheme
    r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
    r.Host = url.Host

    // log.Info().Str("scheme", r.URL.Scheme).Msg("r")

    headers, h_err := json.Marshal(r.Header)

    if h_err != nil {
        log.Error().Err(h_err).Str("req", id.String()).Msg("Could not Marshal Req Headers")
    }

    // it's nice to see when a requests effectively starts- just in case something happens...
    log.Info().RawJSON("headers", headers).Str("uri", r.URL.String()).Str("req", id.String()).Msg("req")

    // intercept things not to change them, just to log them!
    proxy.ModifyResponse = func(res *http.Response) error {
        headers, h_err := json.Marshal(res.Header)

        if h_err != nil {
            log.Error().Err(h_err).Str("req", id.String()).Msg("Could not Marshal Req Headers")
        }

        log.Info().RawJSON("headers", headers).Str("req", id.String()).Msg("res")
        return nil
    }

    proxy.ServeHTTP(w, r)

    latency := time.Since(start).Seconds()
    log.Info().Float64("latency", latency).Str("req", id.String()).Msg("")
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

    zerolog.TimestampFieldName = "t"
    zerolog.LevelFieldName = "l"
    zerolog.MessageFieldName = "m"

    // read basic rules from files on disk? badger?
    // subscribe to "ticker" for new rules - how does base get made? does it reset the ticker data?
}

func main() {
    http.HandleFunc("/", handler)
    // http.HandleFunc("/__iw__/api/add_rule", handler)

    port := fmt.Sprintf(":%s", viper.GetString("port"))
    log.Fatal().Err(http.ListenAndServe(port, nil))
}
