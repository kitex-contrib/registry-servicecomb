package registry

import (
	"errors"
	"fmt"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	"github.com/kitex-contrib/registry-servicecomb/servicecomb"
)

type options struct {
	appId      string
	version    string
	hostName   string
	serviceId  string
	instanceId string
}

// Option is ServiceComb option.
type Option func(o *options)

func WithAppId(appId string) Option {
	return func(o *options) {
		o.appId = appId
	}
}

func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

func WithHostName(hostName string) Option {
	return func(o *options) {
		o.hostName = hostName
	}
}

type serviceCombRegistry struct {
	cli  *sc.Client
	opts options
}

func NewDefaultServiceCombRegistry(opts ...Option) (registry.Registry, error) {
	client, err := servicecomb.NewDefaultServiceCombClient()
	if err != nil {
		return nil, err
	}
	return NewServiceCombRegistry(client, opts...), nil
}

func NewServiceCombRegistry(client *sc.Client, opts ...Option) registry.Registry {
	op := options{
		appId:   "DEFAULT",
		version: "DEFAULT",
	}
	for _, opt := range opts {
		opt(&op)
	}
	return &serviceCombRegistry{cli: client, opts: op}
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
		AppId:       scr.opts.appId,
		Version:     scr.opts.version,
		Status:      "UP",
	})
	if err != nil {
		return fmt.Errorf("register service error: %w", err)
	}
	scr.opts.serviceId = serviceId

	instanceId, err := scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:  serviceId,
		Endpoints:  []string{info.Addr.String()},
		HostName:   scr.opts.hostName,
		Status:     sc.MSInstanceUP,
		Properties: info.Tags,
	})
	if err != nil {
		return fmt.Errorf("register service instance error: %w", err)
	}
	scr.opts.instanceId = instanceId

	return nil
}

func (scr *serviceCombRegistry) Deregister(info *registry.Info) error {
	_, err := scr.cli.UnregisterMicroServiceInstance(scr.opts.serviceId, scr.opts.instanceId)
	if err != nil {
		return fmt.Errorf("Deregister service instance error: %w", err)
	}

	_, err = scr.cli.UnregisterMicroService(scr.opts.serviceId)
	if err != nil {
		return fmt.Errorf("Deregister service error: %w", err)
	}

	return nil
}
