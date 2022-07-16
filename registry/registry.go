package registry

import (
	"errors"
	"fmt"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
)

type serviceCombRegistry struct {
	cli sc.Client
}

func (scr *serviceCombRegistry) Register(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}

	serviceId, err := scr.cli.RegisterService(&discovery.MicroService{
		ServiceName: info.ServiceName,
		AppId:       info.ServiceName,
		Status:      "UP",
	})
	if err != nil {
		return fmt.Errorf("register service error: %w", err)
	}

	_, err = scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:  serviceId,
		Endpoints:  []string{info.Addr.String()},
		Status:     "UP",
		Properties: info.Tags,
	})
	if err != nil {
		return fmt.Errorf("register service instance error: %w", err)
	}

	return nil
}
