package servicecomb

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"os"
	"strconv"
)

const (
	SC_ENV_SERVER_ADDR     = "serverAddr"
	SC_ENV_PORT            = "serverPort"
	SC_DEFAULT_SERVER_ADDR = "127.0.0.1"
	SC_DEFAULT_PORT        = 30100
)

// SCPort Get ServiceComb port from environment variables
func SCPort() int64 {
	portText := os.Getenv(SC_ENV_PORT)
	if len(portText) == 0 {
		return SC_DEFAULT_PORT
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		klog.Errorf("ParseInt failed,err:%s", err.Error())
		return SC_DEFAULT_PORT
	}
	return port
}

// SCAddr Get ServiceComb addr from environment variables
func SCAddr() string {
	addr := os.Getenv(SC_ENV_SERVER_ADDR)
	if len(addr) == 0 {
		return SC_DEFAULT_SERVER_ADDR
	}
	return addr
}
