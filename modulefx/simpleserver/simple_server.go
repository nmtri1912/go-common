package simpleserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nmtri1912/go-common/utils/httputils"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func RunServer(lifecycle fx.Lifecycle) {
	mux := httputils.NewMuxServer(nil)

	port := viper.GetInt("server.port")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Println("HTTP server starting at port:", port)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("HTTP server start error: ", err)
				}
			}()
			return nil
		},
		OnStop: func(c context.Context) error {
			log.Println("HTTP server Shutting down...")
			timeout := viper.GetInt("server.shutdown-timeout-sec")
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	})
}
