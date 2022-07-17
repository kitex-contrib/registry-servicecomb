package reslover

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/go-chassis/sc-client"
	"github.com/kitex-contrib/registry-servicecomb/servicecomb"
)

type options struct {
	appID       string
	versionRule string
	consumerID  string
}

// Option is service-comb resolver option.
type Option func(o *options)

// WithAppID with appID option.
func WithAppID(appID string) Option {
	return func(o *options) { o.appID = appID }
}

// WithVersionRule with versionRule option.
func WithVersionRule(versionRule string) Option {
	return func(o *options) { o.versionRule = versionRule }
}

// WithConsumerID with consumerID option.
func WithConsumerID(consumerID string) Option {
	return func(o *options) { o.consumerID = consumerID }
}

type serviceCombResolver struct {
	cli  sc.Client
	opts options
}

func NewDefaultSCResolver(opts ...Option) (discovery.Resolver, error) {
	client, err := servicecomb.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	return NewSCResolver(*client, opts...), nil
}

func NewSCResolver(cli sc.Client, opts ...Option) discovery.Resolver {
	op := options{
		appID:       "DEFAULT",
		versionRule: "1.0.0",
		consumerID:  "DEFAULT",
	}
	for _, option := range opts {
		option(&op)
	}
	return &serviceCombResolver{
		cli:  cli,
		opts: op,
	}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (scr *serviceCombResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (scr *serviceCombResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := scr.cli.FindMicroServiceInstances(scr.opts.consumerID, scr.opts.appID, desc, scr.opts.versionRule)
	if err != nil {
		return discovery.Result{}, err
	}
	if len(res) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		if in.Status != "UP" {
			continue
		}
		for _, endPoint := range in.Endpoints {
			instances = append(instances, discovery.NewInstance(
				"tcp",
				endPoint,
				10,
				in.Properties))
		}
	}
	if len(instances) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

// Diff computes the difference between two results.
func (scr *serviceCombResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (scr *serviceCombResolver) Name() string {
	return "service-comb-resolver"
}
