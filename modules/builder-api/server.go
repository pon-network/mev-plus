package builderapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/pon-network/mev-plus/modules/builder-api/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type BuilderApiService struct {
	log        *logrus.Entry
	srv        *http.Server
	coreClient *coreCommon.Client

	cfg config.BuilderApiConfig
}

func NewBuilderApiService() *BuilderApiService {

	b := &BuilderApiService{
		log: logrus.NewEntry(logrus.New()),
		cfg: config.BuilderApiConfigDefaults,
	}
	return b
}

func (b *BuilderApiService) CliCommand() *cli.Command {
	return config.NewCommand()
}

func (b *BuilderApiService) Configure(moduleFlags common.ModuleFlags) (err error) {

	for flagName, flagValue := range moduleFlags {
		switch flagName {
		case config.LoggerLevelFlag.Name:
			logLevel, err := logrus.ParseLevel(flagValue)
			if err != nil {
				return err
			}
			b.log.Logger.SetLevel(logLevel)
		case config.LoggerFormatFlag.Name:
			switch flagValue {
			case "json":
				b.log.Logger.SetFormatter(&logrus.JSONFormatter{})
			case "text":
				b.log.Logger.SetFormatter(&logrus.TextFormatter{})
			default:
				return fmt.Errorf("invalid logger format %s", flagValue)
			}
		case config.ListenAddressFlag.Name:
			b.cfg.ListenAddress, err = createUrl(flagValue)
			if err != nil {
				return fmt.Errorf("-%s: invalid url %q", config.ListenAddressFlag.Name, flagValue)
			}
		case config.ServerReadHeaderTimeoutMsFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.ServerReadHeaderTimeoutMs = flagValint
		case config.ServerReadTimeoutMsFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.ServerReadTimeoutMs = flagValint
		case config.ServerWriteTimeoutMsFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.ServerWriteTimeoutMs = flagValint
		case config.ServerIdleTimeoutMsFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.ServerIdleTimeoutMs = flagValint
		case config.ServerMaxHeaderBytesFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.ServerMaxHeaderBytes = flagValint
		default:
			return fmt.Errorf("invalid flag %s", flagName)
		}
	}

	return nil
}

func (b *BuilderApiService) Name() string {
	return config.ModuleName
}

func (b *BuilderApiService) ConnectCore(coreClient *coreCommon.Client, pingId string) error {

	// this is the first and only time the client is set and doesnt need a mutex
	b.coreClient = coreClient

	// test a ping to the core server
	err := b.coreClient.Ping(pingId)
	if err != nil {
		return err
	}

	return nil
}

func (b *BuilderApiService) getRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc(pathRoot, b.handleRoot).Methods(http.MethodGet)
	r.HandleFunc(pathStatus, b.handleStatus).Methods(http.MethodGet)
	r.HandleFunc(pathRegisterValidator, b.handleRegisterValidator).Methods(http.MethodPost)
	r.HandleFunc(pathGetHeader, b.handleGetHeader).Methods(http.MethodGet)
	r.HandleFunc(pathGetPayload, b.handleGetPayload).Methods(http.MethodPost)

	r.Use(mux.CORSMethodMiddleware(r))
	loggedRouter := LoggingMiddleware(b.log, r)
	return loggedRouter
}

func (b *BuilderApiService) Start() (err error) {

	if b.srv != nil {
		return errServerAlreadyRunning
	}

	b.srv = &http.Server{
		Addr:    b.cfg.ListenAddress.Host,
		Handler: b.getRouter(),

		ReadTimeout:       time.Duration(b.cfg.ServerReadTimeoutMs) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(b.cfg.ServerReadHeaderTimeoutMs) * time.Millisecond,
		WriteTimeout:      time.Duration(b.cfg.ServerWriteTimeoutMs) * time.Millisecond,
		IdleTimeout:       time.Duration(b.cfg.ServerIdleTimeoutMs) * time.Millisecond,

		MaxHeaderBytes: b.cfg.ServerMaxHeaderBytes,
	}

	go func ()  {
		listenErr := b.srv.ListenAndServe()
		if listenErr != nil {
			b.log.WithError(listenErr).Error("Failed to start Builder API server")
			err = listenErr
		}
	} ()

	b.log.WithField("listenAddr", b.cfg.ListenAddress.String()).Info("Started Builder API server")

	return nil
}

func (b *BuilderApiService) ListenAddress() string {
	return b.cfg.ListenAddress.String()
}

func (b *BuilderApiService) Stop() error {
	if b.srv == nil {
		return nil
	}

	err := b.srv.Close()
	if err != nil {
		return err
	}

	b.srv = nil

	return nil
}
