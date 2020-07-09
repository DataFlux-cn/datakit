package io

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	// "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/druid"
)

func HTTPServer() {

	mux := http.NewServeMux()
	// mux.HandleFunc("/druid", druid.Handle)

	srv := &http.Server{
		Addr:         config.Cfg.MainCfg.HTTPServerAddr,
		Handler:      requestLogger(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	l.Infof("start http server on %s ok", config.Cfg.MainCfg.HTTPServerAddr)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Error(err)
		}

		l.Info("http server exit")
	}()

	<-datakit.Exit.Wait()
	l.Info("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())

	} else {
		l.Info("http server shutdown ok")
	}

	return
}

type logFormatterParams struct {
	timeStamp  time.Time
	statusCode int
	latency    time.Duration
	clientIP   string
	method     string
	path       string
}

func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var param logFormatterParams

		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		targetMux.ServeHTTP(w, r)

		param.timeStamp = time.Now()
		param.latency = param.timeStamp.Sub(start)

		param.clientIP = r.RemoteAddr
		param.method = r.Method
		// "Status: 200 OK"
		param.statusCode = int(reflect.ValueOf(w).Elem().FieldByName("status").Int())

		if raw != "" {
			path = path + "?" + raw
		}
		param.path = path

		l.Debugf("[HTTP] %3d | %13v | %15s | %-7s  %#v\n",
			param.statusCode,
			param.latency,
			param.clientIP,
			param.method,
			param.path,
		)
	})
}
