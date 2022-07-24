package registry

import (
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

const (
	ServiceName = "demo.kitex-contrib.local"
	AppId       = "DEFAULT"
	Version     = "1.0.0"
	HostName    = "DEFAULT"
)

func getSCClient() (*sc.Client, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: []string{"127.0.0.1:30100"},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func TestNewDefaultSCRegistry(t *testing.T) {
	client, err := getSCClient()
	if err != nil {
		t.Errorf("err:%v", err)
	}
	got := NewSCRegistry(client, WithAppId(AppId), WithVersionRule(Version))
	assert.NotNil(t, got)
}

//  test registry a service
func TestSCRegistryRegister(t *testing.T) {
	client, err := getSCClient()
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
				ServiceName: ServiceName,
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCRegistry(tt.fields.cli, WithAppId(AppId), WithVersionRule(Version), WithHostName(HostName))
			if err := n.Register(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// test deregister a service
func TestSCRegistryDeregister(t *testing.T) {
	client, err := getSCClient()
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
				ServiceName: ServiceName,
			}},
			fields:  fields{client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCRegistry(tt.fields.cli, WithAppId(AppId), WithVersionRule(Version), WithHostName(HostName))
			if err := n.Deregister(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//  test heartbeat
func TestSCRegistryHeartBeat(t *testing.T) {
	client, err := getSCClient()
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
	addr := net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000}
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
				ServiceName: ServiceName,
				Addr:        &addr,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCRegistry(tt.fields.cli, WithAppId(AppId), WithVersionRule(Version), WithHostName(HostName), WithHeartbeatInterval(60))
			if err := n.Register(tt.args.info); err != nil {
				t.Errorf("Register() error = %v", err)
			}
			time.Sleep(time.Minute * 2)
			instances, err := client.FindMicroServiceInstances("", AppId, ServiceName, Version)
			assert.Nil(t, err)
			exist := false
			for _, instance := range instances {
				if funk.ContainsString(instance.Endpoints, addr.String()) {
					exist = true
				}
			}
			assert.True(t, exist)
			_ = n.Deregister(tt.args.info)
		})
	}
}

func TestSCMultipleInstances(t *testing.T) {
	client, err := getSCClient()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	got := NewSCRegistry(client, WithAppId(AppId), WithVersionRule(Version), WithHostName(HostName))
	if !assert.NotNil(t, got) {
		t.Errorf("err: new registry fail")
		return
	}
	time.Sleep(time.Second)

	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081},
	})
	assert.Nil(t, err)
	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8082},
	})
	assert.Nil(t, err)
	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	//time.Sleep(time.Second)
	//instances, err := client.FindMicroServiceInstances("", AppId, ServiceName, Version)
	//assert.Nil(t, err)
	//assert.Equal(t, 3, len(instances), "successful register not three")

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
		Tags: map[string]string{
			"app_id":  AppId,
			"version": Version,
		},
	})
	assert.Nil(t, err)
	time.Sleep(time.Second)
	instances, err := client.FindMicroServiceInstances("", AppId, ServiceName, Version)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(instances), "deregister one, instances num should be two")
}

func existService(t *testing.T, addr net.TCPAddr, wantExist bool) bool {
	req, _ := http.NewRequest("GET", "http://127.0.0.1:30100/registry/v3/instances?appId=DEFAULT&serviceName="+serviceName+"&version=latest", nil)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Errorf("http error: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respByte, _ := ioutil.ReadAll(resp.Body)
	respMap := make(map[string]interface{})
	instances := make([]discovery.MicroServiceInstance, 0)

	err = jsoniter.Unmarshal(respByte, &respMap)
	if instanceList, ok := respMap["instances"]; ok {
		instanceListJsonByte, _ := jsoniter.Marshal(instanceList)
		err = jsoniter.Unmarshal(instanceListJsonByte, &instances)
	}
	if err != nil {
		t.Errorf("jsoniter error: %v\n", err)
	}

	for _, instance := range instances {
		if wantExist && funk.ContainsString(instance.Endpoints, addr.String()) {
			return true
		}
	}
	return !wantExist
}

