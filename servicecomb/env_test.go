package servicecomb

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

// TestEnvFunc test env func
func TestEnvFunc(t *testing.T) {
	assert.Equal(t, "127.0.0.1:30100", SCAddr()+":"+strconv.FormatInt(SCPort(), 10))
}
