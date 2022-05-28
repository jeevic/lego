package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jeevic/lego/components/config"
)

func TestApplication_GetConfig(t *testing.T) {
	cf, err := App.GetConfig()
	assert.Equal(t, cf, (*config.Config)(nil), "config need equal nil")
	assert.NotEqual(t, err, nil, "not init config")
}

func TestApplication_SetTimeLocation(t *testing.T) {
	timeZone := "Asia/Shanghai"

	err := App.SetTimeLocation(timeZone)
	assert.Equal(t, err, nil, "time zone right")
	s := time.Now().In(App.GetTimeLocation()).Format("2006-01-02 15:04:05")
	t.Logf("time:%s", s)
}
