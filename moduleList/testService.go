package modulelist

import (
	"fmt"

	"github.com/pon-pbs/mev-plus/common"
	coreCommon "github.com/pon-pbs/mev-plus/core/common"
)

const name = "TestService"

type TestService struct {
	coreClient *coreCommon.Client
}

func NewTestService() *TestService {

	// Other initializations here specific to your service

	return &TestService{}

}

func (s *TestService) Name() string {
	return name
}

func (s *TestService) Start() error {
	fmt.Println("Test service called to start")
	return nil
}

func (s *TestService) Stop() error {
	fmt.Println("Test service called to stop")
	return nil
}

func (s *TestService) ConnectCore(coreClient *coreCommon.Client, pingId string) error {
	return nil
}

func (s *TestService) Configure(moduleFlags common.ModuleFlags) error {
	fmt.Println("Test service called to configure")
	return nil
}
