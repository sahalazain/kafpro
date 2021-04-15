package api

import (
	"context"
	"os"
	"os/signal"

	"github.com/sahalazain/go-common/config"
	"github.com/sahalazain/kafpro/api/handler"
	"github.com/sahalazain/kafpro/service"
	"github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"github.com/newrelic/opentelemetry-exporter-go/newrelic"
	fotel "github.com/psmarcin/fiber-opentelemetry/pkg/fiber-otel"
	log "github.com/sahalazain/go-common/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

//Server API server
type Server interface {
	Configure(ctx context.Context, config config.Getter) error
	Serve() error
}

//FiberServer gofiber api server
type FiberServer struct {
	App     *fiber.App
	Config  config.Getter
	Service service.App
}

var l *logrus.Entry

// NewFiberServer create instance of fiber server
func NewFiberServer() *FiberServer {

	return &FiberServer{
		App: fiber.New(),
	}
}

//Configure configure fiber server
func (f *FiberServer) Configure(ctx context.Context, config config.Getter) error {
	l = log.GetLoggerContext(ctx, "api", "Configure")
	f.Config = config

	svc, err := service.GetService(config)
	if err != nil {
		return err
	}

	f.Service = svc

	if err := f.registerHandlers(); err != nil {
		l.WithError(err).Error("Error registering handler")
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		f.App.Shutdown()
	}()
	return nil
}

//Serve start fiber server
func (f *FiberServer) Serve() error {
	f.App.Listen(f.Config.GetString("address") + ":" + f.Config.GetString("port"))
	return nil
}

func (f *FiberServer) registerHandlers() error {

	f.App.Use(recover.New())
	f.App.Use(logger.New())

	tracer := f.getTracerProvider().Tracer(f.Config.GetString("name"))

	f.App.Use(fotel.New(fotel.Config{
		Tracer: tracer,
	}))

	f.App.Get("/status", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	f.App.Get("/topic/:topic/:group", handler.Get(f.Service))
	f.App.Post("/topic/:topic", handler.Send(f.Service))
	f.App.Post("/topic/:topic/:key", handler.Send(f.Service))

	return nil
}

func (f *FiberServer) getTracerProvider() trace.TracerProvider {
	var provider trace.TracerProvider
	switch f.Config.GetString("tracer") {
	case "newrelic":
		l.Info("Using new relic tracer")
		exporter, err := newrelic.NewExporter(
			"My Service", f.Config.GetString("newrelic_apikey"),
			telemetry.ConfigSpansURLOverride("https://nr-internal.aws-us-east-1.tracing.edge.nr-data.net/trace/v1"),
		)
		if err != nil {
			break
		}
		provider = sdk.NewTracerProvider(sdk.WithSyncer(exporter))
	case "stdout":
		l.Info("Using stdout tracer")
		otExporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
		if err != nil {
			break
		}
		provider = sdk.NewTracerProvider(sdk.WithSyncer(otExporter))
	case "jaeger":
		l.Info("Using jaeger tracer")
		pr, flush, err := jaeger.NewExportPipeline(
			jaeger.WithCollectorEndpoint(f.Config.GetString("jaeger_url")),
			jaeger.WithProcess(jaeger.Process{
				ServiceName: f.Config.GetString("name"),
				Tags: []label.KeyValue{
					label.String("exporter", "jaeger"),
				},
			}),
			jaeger.WithSDK(&sdk.Config{DefaultSampler: sdk.AlwaysSample()}),
		)
		if err != nil {
			break
		}
		provider = pr
		defer flush()
	default:
		l.Info("Using no op tracer")
		provider = trace.NewNoopTracerProvider()
	}

	if provider == nil {
		l.Info("Using no op tracer")
		provider = trace.NewNoopTracerProvider()
	}

	otel.SetTracerProvider(provider)

	return provider
}
