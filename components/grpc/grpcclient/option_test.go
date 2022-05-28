package grpcclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewDefaultOptions(t *testing.T) {
	opts := NewDefaultOptions()
	assert.Equal(t, opts.PoolCap, defaultClientPoolCap)
}

func TestWithPoolCap(t *testing.T) {
	opts := NewOptions(WithPoolCap(1))
	assert.Equal(t, opts.PoolCap, int64(1))
}

func TestWithDailTimeOut(t *testing.T) {
	opts := NewOptions(WithDailTimeOut(time.Second))
	assert.Equal(t, opts.DialTimeout, time.Second)
}

func TestWithInsecure(t *testing.T) {
	opts := NewOptions(WithInsecure(true))
	assert.Equal(t, opts.Insecure, true)
}

func TestWithKeepAlive(t *testing.T) {
	opts := NewOptions(WithKeepAlive(time.Second))
	assert.Equal(t, opts.KeepAlive, time.Second)
}

func TestWithKeepAliveTimeout(t *testing.T) {
	opts := NewOptions(WithKeepAliveTimeout(time.Second))
	assert.Equal(t, opts.KeepAliveTimeout, time.Second)
}

func TestWithKeepAlivePermitWithoutStream(t *testing.T) {
	opts := NewOptions(WithKeepAlivePermitWithoutStream(true))
	assert.Equal(t, opts.KeepAlivePermitWithoutStream, true)
}

func TestWithCredentials(t *testing.T) {
	no := insecure.NewCredentials()
	opts := NewOptions(WithCredentials(no))
	assert.Equal(t, opts.Credentials, no)
}
