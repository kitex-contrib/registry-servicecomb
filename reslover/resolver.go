package reslover

import (
	"context"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/go-chassis/sc-client"
	"github.com/kitex-contrib/registry-servicecomb/servicecomb"
)

type serviceCombResolver struct {
	cli sc.Client
}

func NewDefaultSCResolver() (discovery.Resolver, error) {
	client, err := servicecomb.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	return NewSCResolver(*client), nil
}

func NewSCResolver(cli sc.Client) discovery.Resolver {
	return &serviceCombResolver{
		cli: cli,
	}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (n *serviceCombResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (n *serviceCombResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {

}

// Diff computes the difference between two results.
func (n *serviceCombResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (n *serviceCombResolver) Name() string {
	return "service-comb-resolver"
}
