package cmd

import (
	"context"
	"log"

	"github.com/sahalazain/go-common/config"
	"github.com/sahalazain/go-common/logger"
	"github.com/sahalazain/kafpro/api"
	"github.com/spf13/cobra"
)

var configURL string
var serverCommand = &cobra.Command{
	Use: "serve",
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
	PostRun: func(cmd *cobra.Command, args []string) {

	},
}

func startServer() {
	conf, err := config.Load(DefaultConfig, configURL)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	srv := api.NewFiberServer()
	if err := srv.Configure(ctx, conf); err != nil {
		log.Fatal(err)
	}

	logger.Configure(conf)

	srv.Serve()
}
