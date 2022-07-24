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

const serviceName = "demo.kitex-contrib.local"

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
	got := NewServiceCombRegistry(client, WithAppId("DEFAULT"), WithVersionRule("1.0.0"))
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
				ServiceName: serviceName,
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewServiceCombRegistry(tt.fields.cli, WithAppId("DEFAULT"), WithVersionRule("1.0.0"), WithHostName("DEFAULT"))
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
				ServiceName: serviceName,
			}},
			fields:  fields{client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewServiceCombRegistry(tt.fields.cli, WithAppId("DEFAULT"), WithVersionRule("1.0.0"), WithHostName("DEFAULT"))
			if err := n.Deregister(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//  test registry a service
func TestServiceCombRegistryHeartBeat(t *testing.T) {
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
	addr := net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "common-not-keepalive",
			fields: fields{client},
			args: args{info: &registry.Info{
				ServiceName: serviceName,
				Addr:        &addr,
			}},
			wantErr: false,
		},
		{
			name:   "common-keepalive",
			fields: fields{client},
			args: args{info: &registry.Info{
				ServiceName: serviceName,
				Addr:        &addr,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alive := false
			if tt.name == "common-keepalive" {
				alive = true
			}
			n := NewServiceCombRegistry(tt.fields.cli, WithAppId("DEFAULT"), WithVersionRule("1.0.0"), WithHostName("DEFAULT"), WithKeepAlive(alive), WithHeartbeatInterval(60))
			if err := n.Register(tt.args.info); err != nil {
				t.Errorf("Register() error = %v", err)
			}
			time.Sleep(time.Minute * 3)
			assert.True(t, existService(t, addr, alive))
			_ = n.Deregister(tt.args.info)
			time.Sleep(time.Second * 10)
		})
	}
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
