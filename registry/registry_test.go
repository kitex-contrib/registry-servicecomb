package registry

import (
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/sc-client"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func getServiceCombClient() (*sc.Client, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: []string{"127.0.0.1:30100"},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func TestNewDefaultServiceCombRegistry(t *testing.T) {
	client, err := getServiceCombClient()
	if err != nil {
		t.Errorf("err:%v", err)
	}
	got := NewServiceCombRegistry(client, WithAppId("DEFAULT"), WithVersion("DEFAULT_GROUP"))
	assert.NotNil(t, got)
}

//  test registry a service
func TestServiceCombRegistryRegister(t *testing.T) {
	client, err := getServiceCombClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli *sc.Client
	}
	type args struct {
		info *registry.Info
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "common",
			fields: fields{client},
			args: args{info: &registry.Info{
				ServiceName: "demo.kitex-contrib.local",
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000},
				Weight:      999,
				StartTime:   time.Now(),
				Tags:        map[string]string{"env": "local"},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewServiceCombRegistry(tt.fields.cli, WithAppId("DEFAULT"), WithVersion("0.1"), WithHostName("ServiceComb-Test"), WithServiceId("TestServiceId"))
			if err := n.Register(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// test deregister a service
func TestServiceCombRegistryDeregister(t *testing.T) {
	client, err := getServiceCombClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli *sc.Client
	}
	type args struct {
		info *registry.Info
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "common",
			args: args{info: &registry.Info{
				ServiceName: "demo.kitex-contrib.local",
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 300},
				Weight:      999,
				StartTime:   time.Now(),
				Tags:        map[string]string{"env": "local"},
			}},
			fields:  fields{client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewServiceCombRegistry(tt.fields.cli, WithAppId("DEFAULT"), WithVersion("0.1"), WithHostName("ServiceComb-Test"), WithServiceId("TestServiceId"))
			if err := n.Deregister(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
