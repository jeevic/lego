package util

import (
	"testing"
)

func TestContain(t *testing.T) {
	var obj = "xml"
	var target = []string{"json", "yaml", "xml", "ini"}

	b, e := Contain(obj, target)

	if b != true || e != nil {
		t.Error("contain slice 应在slice中,但未在", obj, target, b, e)
	}

	obj = "aaaa"
	b, e = Contain(obj, target)
	if b == true || e == nil {
		t.Error("contain slice 不应该slice中,但在", obj, target, b, e)
	}

	obj = "yaml"
	targetArr := [5]string{"json", "yaml", "xml", "ini"}
	b, e = Contain(obj, targetArr)
	if b != true || e != nil {
		t.Error("contain array 应在array中,但未在", obj, targetArr, b, e)
	}

	obj = "bbbb"
	b, e = Contain(obj, targetArr)
	if b == true || e == nil {
		t.Error("contain array array,但在", obj, targetArr, b, e)
	}

	var strMap = map[string]string{
		"json": "json",
		"yaml": "yaml",
		"xml":  "xml",
		"ini":  "ini",
	}

	obj = "xml"
	b, e = Contain(obj, strMap)
	if b != true || e != nil {
		t.Error("contain string map 应在map中,但未在", obj, strMap, b, e)
	}

	obj = "bbbb"
	b, e = Contain(obj, strMap)
	if b == true || e == nil {
		t.Error("contain string map 不应在,但在", obj, strMap, b, e)
	}

}

func BenchmarkContain(b *testing.B) {
	b.StopTimer()

	b.StartTimer()
	var target = [...]string{"json", "yaml", "xml", "ini"}
	for i := 0; i < b.N; i++ {
		isBool, err := Contain("json", target)

		if isBool == false || err != nil {
			b.Error("contain array 应该包含 但未包含", "json", target)
		}
	}
}

func TestGetLocalIp(t *testing.T) {
	ip, _ := GetLocalIp()
	t.Log("local ip:", ip)
	if len(ip) < 0 {
		t.Error("get local ip error!")
	}
}
