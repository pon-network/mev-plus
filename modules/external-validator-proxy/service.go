package externalvalidatorproxy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/pon-network/mev-plus/modules/external-validator-proxy/config"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

type ExternalValidatorProxyService struct {
	httpClient http.Client
	log        *logrus.Entry
	coreClient *coreCommon.Client

	cfg config.ProxyConfig
}

func NewExternalValidatorProxyService() *ExternalValidatorProxyService {

	p := &ExternalValidatorProxyService{
		log: logrus.NewEntry(logrus.New()).WithField("moduleExecution", config.ModuleName),
		cfg: config.ProxyConfigDefaults,
		httpClient: http.Client{
			Timeout: time.Duration(config.ProxyConfigDefaults.RequestTimeoutMs) * time.Millisecond,
		},
	}
	return p
}

func (p *ExternalValidatorProxyService) Configure(moduleFlags common.ModuleFlags) (err error) {

	for flagName, flagValue := range moduleFlags {
		switch flagName {
		case config.LoggerLevelFlag.Name:
			logLevel, err := logrus.ParseLevel(flagValue)
			if err != nil {
				return err
			}
			p.log.Logger.SetLevel(logLevel)
		case config.LoggerFormatFlag.Name:
			switch flagValue {
			case "json":
				p.log.Logger.SetFormatter(&logrus.JSONFormatter{})
			case "text":
				p.log.Logger.SetFormatter(&logrus.TextFormatter{})
			default:
				return fmt.Errorf("invalid logger format %s", flagValue)
			}
		case config.AddressFlag.Name:
			p.cfg.Address, err = createUrl(flagValue)
			if err != nil {
				return fmt.Errorf("-%s: invalid url %q", config.AddressFlag.Name, flagValue)
			}
		case config.RequestTimeoutMsFlag.Name:
			requestTimeoutMs, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				return err
			}
			p.cfg.RequestTimeoutMs = int(requestTimeoutMs)
			p.httpClient.Timeout = time.Duration(p.cfg.RequestTimeoutMs) * time.Millisecond
		case config.RequestMaxRetriesFlag.Name:
			requestMaxRetries, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				return err
			}
			p.cfg.RequestMaxRetries = int(requestMaxRetries)
		default:
			return fmt.Errorf("invalid flag %s", flagName)
		}
	}

	return nil
}

func (p *ExternalValidatorProxyService) Name() string {
	return config.ModuleName
}

func (p *ExternalValidatorProxyService) ConnectCore(coreClient *coreCommon.Client, pingId string) error {

	// this is the first and only time the client is set and doesnt need a mutex
	p.coreClient = coreClient

	// test a ping to the core server
	err := p.coreClient.Ping(pingId)
	if err != nil {
		return err
	}

	return nil
}

func (p *ExternalValidatorProxyService) Start() error {

	if p.cfg.Address == nil {
		// No address set, do not start
		return nil
	}

	var builderApiAddress string
	err := p.coreClient.Call(&builderApiAddress, "builderApi_listenAddress", false, nil)
	if err != nil {
		return err
	}

	if builderApiAddress == p.cfg.Address.String() {
		return fmt.Errorf("proxy address %s is the same as the builder api address %s", p.cfg.Address.String(), builderApiAddress)
	}

	// Connect to the block aggregator module
	ctx := context.Background()
	err = p.coreClient.Notify(ctx, "blockAggregator_connectBlockSource", false, nil, config.ModuleName)
	if err != nil {
		return err
	}

	p.log.WithField("proxyAddress", p.cfg.Address.String()).Info("Started External Validator Proxy service")

	return nil
}

func (p *ExternalValidatorProxyService) Stop() error {
	return nil
}

func NewCommand() *cli.Command {
	return config.NewCommand()
}
