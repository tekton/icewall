package main

import (
    "fmt"
    // "net"
    "net/http"
    "net/url"
    "net/http/httputil"
    "encoding/json"
    "time"
    "os"
    "io/ioutil"
    // "context"

    //
    "github.com/rs/xid"
    "github.com/spf13/viper"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    //
    // "github.com/go-redis/redis/v8"
)

// var ctx = context.Background()
// var redisConn *redis.Client

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // laod balancer health short circuit rule check
    if (r.URL.Path == viper.GetString("health_check.path")) {
        log.Debug().Msg("health check")
        // we have a health check to perform! Now, which kind?
        if (viper.GetString("health_check.action") == "file") {
            // ah, a file- we should serve it back and exit early!
            w.Header().Set("Content-Type", viper.GetString("health_check.type"))

            // what if we want it to say its in maintenance mode?
            if (viper.GetString("health_check.maintenance.file") != "") {
                // open the file, read the contents...
                maintenanceFile, err := os.Open(viper.GetString("health_check.maintenance.file"))
                if err != nil {
                    log.Error().Err(err)
                } else { // yes, and else, sue me for not doing a function
                    defer maintenanceFile.Close()
                    maintenanceData, readErr := ioutil.ReadAll(maintenanceFile)
                    if readErr != nil {
                        log.Error().Err(readErr)
                        // not going to set the status code to a 5XX because it isn't that opinionated
                    } else {
                        if (string(maintenanceData) == viper.GetString("health_check.maintenance.check_val")) {
                            log.Info().Msg("maintenance mode")
                            // set the status code!
                            w.WriteHeader(viper.GetInt("health_check.maintenance.status_code"))
                        }
                    }
                }
            }

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

    if (viper.GetBool("rules_enabled") == true) {
        if (viper.GetBool("global_throttle.enabled") == true) {
            // general check first!
        }
    }

    // header rules
    // path rules
    // query string rules

    // headers to look for as an array?

    url, _ := url.Parse(host)
    proxy := httputil.NewSingleHostReverseProxy(url)
    // make a director to control timeouts...

    // fmt.Printf("%#v", proxy.Transport)

    // TODO: add settings with defaults...the base versions of the dial are fine though...
    // proxy.Transport = &http.Transport{
    //     DialContext: (&net.Dialer{
    //         Timeout:   1 * time.Second,
    //         KeepAlive: 1 * time.Second,
    //         DualStack: true,
    //     }).DialContext,
    // }

    // fmt.Println(" ")
    // fmt.Printf("%#v", proxy.Transport)
    // fmt.Println(" ")

    // defer proxy.Close()

    r.URL.Host = url.Host
    r.URL.Scheme = url.Scheme
    if (viper.GetBool("forwarded_host") == true) {
        if r.Header.Get("Host") != "" && r.Header.Get("Host") != "::1" {
            r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
        }
    }
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

        // TODecide readd x-iw-id just in case it got dropped?

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
    viper.SetDefault("forwarded_host", true)

    viper_err := viper.ReadInConfig()   // Find config, read config, or else...
    if viper_err != nil {
        panic(fmt.Errorf("Fatal error config file: %s \n", viper_err))
    }

    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    if viper.GetString("log_level") == "debug" {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    }

    zerolog.TimestampFieldName = "t"
    zerolog.LevelFieldName = "l"
    zerolog.MessageFieldName = "m"

    log.Info().Msg("starting icewall")

    // read basic rules from files on disk? badger?
    // subscribe to "ticker" for new rules - how does base get made? does it reset the ticker data?
}

func main() {
    http.HandleFunc("/", handler)
    // http.HandleFunc("/__iw__/api/add_rule", handler)

    // redisConn = redis.NewClient(&redis.Options{
    //     Network:    "tcp",
    //     Addr:       "127.0.0.1:6379",
    // })
    // defer redisConn.Close()
    // val := redisConn.Ping(ctx)
    // log.Info().Str("val", val.String()).Msg("ping")

    port := fmt.Sprintf(":%s", viper.GetString("port"))
    log.Fatal().Err(http.ListenAndServe(port, nil))
}
