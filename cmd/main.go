package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elliottpolk/stuber"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var (
	version  string
	compiled string = fmt.Sprint(time.Now().Unix())

	cfgFileFlag = &cli.StringFlag{
		Name:    "config-file",
		Aliases: []string{"c", "cfg", "confg", "config"},
		Usage:   "optional path to config file",
	}

	httpPortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "http-port",
		Value:   "8080",
		Usage:   "HTTP port to listen on",
		EnvVars: []string{"STUBER_HTTP_PORT"},
	})

	httpsPortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "tls-port",
		Value:   "8443",
		Usage:   "HTTPS port to listen on",
		EnvVars: []string{"STUBER_HTTPS_PORT"},
	})

	tlsCertFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "tls-cert",
		Usage:   "TLS certificate file for HTTPS",
		EnvVars: []string{"STUBER_TLS_CERT"},
	})

	tlsKeyFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "tls-key",
		Usage:   "TLS key file for HTTPS",
		EnvVars: []string{"STUBER_TLS_KEY"},
	})

	dataDirFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "data-dir",
		Aliases: []string{"d", "dir", "data"},
		Usage:   "data directory for stub JSON files",
		Value:   "./data",
	})
)

func main() {
	ct, err := strconv.ParseInt(compiled, 0, 0)
	if err != nil {
		panic(err)
	}

	app := cli.App{
		Name:        "stuber",
		Description: "Stuber /st(y)o͞o·ber/ is a configurable stubbing tool meant to provide mocked stub services, typically for testing.",
		Copyright:   fmt.Sprintf("Copyright © 2018-%s Elliott Polk", time.Now().Format("2006")),
		Version:     version,
		Compiled:    time.Unix(ct, -1),
		Flags: []cli.Flag{
			cfgFileFlag,
			httpPortFlag,
			httpsPortFlag,
			tlsCertFlag,
			tlsKeyFlag,
			dataDirFlag,
		},
		Before: func(ctx *cli.Context) error {
			if cfg := ctx.String(cfgFileFlag.Name); len(strings.TrimSpace(cfg)) > 0 {
				return altsrc.InitInputSourceWithContext(ctx.Command.Flags, altsrc.NewYamlSourceFromFlagFunc(cfgFileFlag.Name))(ctx)
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {

			var (
				httpPort  = ctx.String(httpPortFlag.Name)
				httpsPort = ctx.String(httpsPortFlag.Name)

				cert = ctx.String(tlsCertFlag.Name)
				key  = ctx.String(tlsKeyFlag.Name)
			)

			mux, err := stuber.LoadRoutes(http.NewServeMux(), ctx.String(dataDirFlag.Name))
			if err != nil {
				return cli.Exit(err, 1)
			}

			// HTTPS
			if len(cert) > 0 && len(key) > 0 {
				go func() {
					if _, err := os.Stat(cert); err != nil {
						log.Error(errors.Wrap(err, "unable to access TLS cert file"))
						return
					}

					if _, err := os.Stat(key); err != nil {
						log.Error(errors.Wrap(err, "unable to access TLS key file"))
						return
					}

					log.Infof("HTTPS listening on port %s", httpsPort)
					log.Fatal(http.ListenAndServeTLS(":"+httpsPort, cert, key, mux))
				}()
			}

			// HTTP
			log.Infof("HTTP listening on port %s", httpPort)
			log.Fatal(http.ListenAndServe(":"+httpPort, mux))

			return nil
		},
	}

	app.Run(os.Args)
}
