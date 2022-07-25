package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	"github.com/kitex-contrib/registry-servicecomb/servicecomb"
	"github.com/thoas/go-funk"
	"time"
)

type options struct {
	appId                   string
	versionRule             string
	hostName                string
	heartbeatIntervalSecond int32
}

// Option is ServiceComb option.
type Option func(o *options)

// WithAppId with app id option
func WithAppId(appId string) Option {
	return func(o *options) {
		o.appId = appId
	}
}

// WithVersionRule with version rule option
func WithVersionRule(versionRule string) Option {
	return func(o *options) {
		o.versionRule = versionRule
	}
}

// WithHostName with host name option
func WithHostName(hostName string) Option {
	return func(o *options) {
		o.hostName = hostName
	}
}

func WithHeartbeatInterval(second int32) Option {
	return func(o *options) {
		o.heartbeatIntervalSecond = second
	}
}

type serviceCombRegistry struct {
	cli  *sc.Client
	opts options
}

// NewDefaultSCRegistry create a new default ServiceComb registry
func NewDefaultSCRegistry(opts ...Option) (registry.Registry, error) {
	client, err := servicecomb.NewDefaultSCClient()
	if err != nil {
		return nil, err
	}
	return NewSCRegistry(client, opts...), nil
}

// NewSCRegistry create a new ServiceComb registry
func NewSCRegistry(client *sc.Client, opts ...Option) registry.Registry {
	op := options{
		appId:       "DEFAULT",
		versionRule: "1.0.0",
	}
	for _, opt := range opts {
		opt(&op)
	}
	return &serviceCombRegistry{cli: client, opts: op}
}

// Register a service info to ServiceComb
func (scr *serviceCombRegistry) Register(info *registry.Info) error {
	ctx := context.Background()
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}

	serviceID, err := scr.cli.RegisterService(&discovery.MicroService{
		ServiceName: info.ServiceName,
		AppId:       scr.opts.appId,
		Version:     scr.opts.versionRule,
		Status:      sc.MSInstanceUP,
	})
	if err != nil {
		return fmt.Errorf("register service error: %w", err)
	}

	healthCheck := &discovery.HealthCheck{
		Mode:     "push",
		Interval: 30,
		Times:    3,
	}
	if scr.opts.heartbeatIntervalSecond > 0 {
		healthCheck.Interval = scr.opts.heartbeatIntervalSecond
	}

	instanceId, err := scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:   serviceID,
		Endpoints:   []string{info.Addr.String()},
		HostName:    scr.opts.hostName,
		HealthCheck: healthCheck,
		Status:      sc.MSInstanceUP,
		Properties:  info.Tags,
	})
	if err != nil {
		return fmt.Errorf("register service instance error: %w", err)
	}

	go func(ctx context.Context, serviceId, instanceId string) {
		defer func() {
			if r := recover(); r != nil {
				klog.CtxErrorf(ctx, "beat to ServerComb panic:%+v", r)
				_ = scr.Deregister(info)
			}
		}()
		ticker := time.NewTicker(time.Second * time.Duration(healthCheck.Interval))
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				success, err := scr.cli.Heartbeat(serviceId, instanceId)
				if err != nil || !success {
					klog.CtxErrorf(ctx, "beat to ServerComb return error:%+v instance:%v", err, instanceId)
					ticker.Stop()
					return
				}
			}
		}
	}(ctx, serviceID, instanceId)

	return nil
}

func (scr *serviceCombRegistry) Deregister(info *registry.Info) error {
	serviceId, err := scr.cli.GetMicroServiceID(scr.opts.appId, info.ServiceName, scr.opts.versionRule, "")
	if err != nil {
		return fmt.Errorf("get service-id error: %w", err)
	}
	if info.Addr == nil {
		_, err = scr.cli.UnregisterMicroService(serviceId)
		if err != nil {
			return fmt.Errorf("deregister service error: %w", err)
		}
	} else {
		instanceId := ""
		instances, err := scr.cli.FindMicroServiceInstances("", info.Tags["app_id"], info.ServiceName, info.Tags["version"])
		if err != nil {
			return fmt.Errorf("get instances error: %w", err)
		}
		for _, instance := range instances {
			if funk.ContainsString(instance.Endpoints, info.Addr.String()) {
				instanceId = instance.InstanceId
			}
		}
		klog.Info(instances)
		if instanceId != "" {
			_, err = scr.cli.UnregisterMicroServiceInstance(serviceId, instanceId)
			if err != nil {
				return fmt.Errorf("deregister service error: %w", err)
			}
		}
	}

	return nil
}
