package reslover

import (
	"context"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/sc-client"
	scregistry "github.com/kitex-contrib/registry-servicecomb/registry"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
	"time"
)

var (
	SCClient      = &sc.Client{}
	ServiceName   = "demo.kitex-contrib.local"
	ServiceAddr   = net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
	AppId         = "DEFAULT"
	Version       = "1.0.0"
	LatestVersion = "latest"
	HostName      = "DEFAULT"
	svcInfo       = &registry.Info{
		ServiceName: ServiceName,
		Addr:        &ServiceAddr,
		Tags: map[string]string{
			"app_id":  AppId,
			"version": LatestVersion,
		},
	}
)

func init() {
	cli, err := getSCClient()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	err = scregistry.NewSCRegistry(cli, scregistry.WithAppId(AppId), scregistry.WithVersionRule(Version), scregistry.WithHostName(HostName)).Register(svcInfo)
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	SCClient = cli
}

func getSCClient() (*sc.Client, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: []string{"127.0.0.1:30100"},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// TestNewDefaultSCResolver test new a default SC resolver
func TestNewDefaultSCResolver(t *testing.T) {
	r, err := NewDefaultSCResolver()
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

// TestSCResolverResolve test Resolve a service
func TestSCResolverResolve(t *testing.T) {
	type fields struct {
		cli sc.Client
	}
	type args struct {
		ctx  context.Context
		desc string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    discovery.Result
		wantErr bool
	}{
		{
			name: "common",
			args: args{
				ctx:  context.Background(),
				desc: ServiceName,
			},
			fields: fields{cli: *SCClient},
		},
		{
			name: "wrong desc",
			args: args{
				ctx:  context.Background(),
				desc: "xxxx.kitex-contrib.local",
			},
			fields:  fields{cli: *SCClient},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCResolver(tt.fields.cli)
			_, err := n.Resolve(tt.args.ctx, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), "micro-service does not exist") {
				t.Errorf("Resolve err is not expectant")
				return
			}
		})
	}

	err := scregistry.NewSCRegistry(SCClient).Deregister(&registry.Info{
		ServiceName: ServiceName,
	})
	if err != nil {
		t.Errorf("Deregister Fail")
		return
	}
}
