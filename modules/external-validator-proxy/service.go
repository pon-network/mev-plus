package externalvalidatorproxy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
			addressList := strings.Split(flagValue, ",")
			if len(addressList) > 2 {
				return fmt.Errorf("-%s: too many addresses provided", config.AddressFlag.Name)
			}
			for _, address := range addressList {
				// check address string is not duplicate by checking if you can find more than one occurence in the flagValue
				if strings.Count(flagValue, address) > 1 {
					return fmt.Errorf("-%s: duplicate external proxy address provided %q", config.AddressFlag.Name, address)
				}
				addressUrl, err := createUrl(address)
				if err != nil {
					return fmt.Errorf("-%s: invalid url %q", config.AddressFlag.Name, address)
				}
				p.cfg.Addresses = append(p.cfg.Addresses, addressUrl)
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

	if len(p.cfg.Addresses) == 0 {
		// No address set, do not start and do not receive notifications from block aggregator
		ctx := context.Background()
		err := p.coreClient.Notify(ctx, "blockAggregator_excludeFromNotifications", false, nil, config.ModuleName)
		if err != nil {
			return err
		}

		return nil
	}

	var builderApiAddress string
	err := p.coreClient.Call(&builderApiAddress, "builderApi_listenAddress", false, nil)
	if err != nil {
		return err
	}

	var joinedAddresses string

	for _, address := range p.cfg.Addresses {
		if builderApiAddress == address.String() {
			return fmt.Errorf("proxy address %s is the same as the builder api address %s", address.String(), builderApiAddress)
		}
		if len(joinedAddresses) > 0 {
			joinedAddresses = strings.Join([]string{joinedAddresses, address.String()}, ",")
		} else {
			joinedAddresses = address.String()
		}
	}

	// Connect to the block aggregator module
	ctx := context.Background()
	err = p.coreClient.Notify(ctx, "blockAggregator_connectBlockSource", false, nil, config.ModuleName)
	if err != nil {
		return err
	}

	p.log.WithField("connectedProxyAddresses", joinedAddresses).Info("Started External Validator Proxy service")

	return nil
}

func (p *ExternalValidatorProxyService) Stop() error {
	return nil
}

func NewCommand() *cli.Command {
	return config.NewCommand()
}
