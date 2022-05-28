package godis

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build(t *testing.T) {
	zkAddr := []string{"10.126.173.11:2181", "10.126.173.12:2181", "10.126.173.13:2181"}
	zkDir := "/jodis/codis-caifeng_test"

	pool, err := Create().SetZookeeperClient(zkAddr, zkDir, 3000).SetDb(5).SetPoolSize(10).Build()

	assert.Equal(t, err, nil)

	cli, _ := pool.GetClient()

	cmd := cli.Set("test_1", 10, 100*time.Second)
	str, err := cmd.Result()
	assert.Equal(t, err, nil)
	t.Log(str)

}

func BenchmarkBuilder_Build(b *testing.B) {
	zkAddr := []string{"10.126.173.11:2181", "10.126.173.12:2181", "10.126.173.13:2181"}
	zkDir := "/jodis/codis-caifeng_test"
	pool, err := Create().SetZookeeperClient(zkAddr, zkDir, 3000).SetDb(5).SetPoolSize(10).Build()
	assert.Equal(b, err, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli, _ := pool.GetClient()
		cmd := cli.Set(fmt.Sprintf("test_%d", i), i, 100*time.Second)
		_, err := cmd.Result()
		assert.Equal(b, err, nil)
	}
}
