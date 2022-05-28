package redis

import (
	"fmt"
	"testing"
)

func TestNewRedisUniversal(t *testing.T) {

	setting := &Setting{
		//Hosts: "10.103.17.53:6385,10.103.17.53:6386,10.103.17.53:6387",
		//MasterName: "mymaster",
		//Hosts:      []string{"10.103.17.53:6390", "10.103.17.53:6391", "10.103.17.53:6392"},
		//Hosts: []string{"10.103.17.53:26390", "10.103.17.53:26391", "10.103.17.53:26392"},
		Hosts: []string{"10.103.17.53:6379", "10.103.17.53:6380", "10.103.17.53:6381", "10.103.17.53:6382", "10.103.17.53:6383", "10.103.17.53:6384"},
	}
	redis := NewRedisUniversal(setting)
	//for i := 0;i<100000;i++ {
	//	redis.Client.Set("A_rank_0_00vCH6X3rIhw"+string(i),1,5*time.Hour)
	//}
	num, err := redis.Client.Get("222222").Int()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(num)
	result := redis.Client.Ping()
	fmt.Println(result)
}
