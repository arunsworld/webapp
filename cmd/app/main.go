package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

//go:embed all:embed/*
var webContent embed.FS

type config struct {
	port      int
	debugMode bool
}

var appName = "demo webapp with Go, TailwindCSS and HTMX"

func main() {

	appFlags, cfg := createFlags()

	app := &cli.App{
		Name:  appName,
		Flags: appFlags,
		Action: func(cCtx *cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if cfg.debugMode {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			return run(ctx, *cfg)
		},
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	log.Info().Msgf("%s starting...", appName)

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msgf("%s terminated...", appName)
}

func createFlags() ([]cli.Flag, *config) {
	cfg := &config{}

	return []cli.Flag{
		&cli.IntFlag{
			Name:        "port",
			Value:       6160,
			EnvVars:     []string{"PORT"},
			Destination: &cfg.port,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "turn on debug mode",
			Aliases:     []string{"d"},
			EnvVars:     []string{"DEBUG"},
			Destination: &cfg.debugMode,
		},
	}, cfg
}

func run(ctx context.Context, cfg config) error {
	webApp := fiber.New(fiber.Config{
		Prefork: false,
		Views:   html.NewFileSystem(http.FS(mustFSSub(webContent, "embed")), ".html"),
	})
	go appShutdownOnCtxCancel(ctx, webApp)

	webApp.Use(logger.New(), etag.New())

	setupStatic(webApp, mustFSSub(webContent, "embed/static"))
	registerAppRoutes(webApp)

	log.Debug().Int("port", cfg.port).Msg("serving webserver...")
	return webApp.Listen(fmt.Sprintf("localhost:%d", cfg.port))
}

func appShutdownOnCtxCancel(ctx context.Context, app *fiber.App) {
	<-ctx.Done()
	if err := app.Shutdown(); err != nil {
		log.Printf("error while shutting down app: %v", err)
	}
}

func mustFSSub(src fs.FS, dir string) fs.FS {
	fsys, err := fs.Sub(src, dir)
	if err != nil {
		panic(err)
	}
	return fsys
}
