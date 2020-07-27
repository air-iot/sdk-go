package ext

import (
	"github.com/air-iot/sdk-go/traefik"
	"testing"
)

func init() {
	traefik.Host = "iot.tmis.top"
	traefik.Port = 8010
	traefik.Enable = true
	traefik.AppKey = "b9bd592b-2d79-4f5c-d583-aad18ebe00ca"
	traefik.AppSecret = "c5de1068-79fd-b32b-a4f8-291c337111fa"
}

func TestModelClient_FindById(t *testing.T) {
	cli := NewExtClient("新表")
	var r = make(map[string]interface{})
	err := cli.FindById("5f1aa80eac624e29f1678fd3", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestExtClient_SaveMany(t *testing.T) {
	cli := NewExtClient("新表")
	var r = make(map[string]interface{}, 0)
	dataMap := []map[string]interface{}{
		{
			"boolean-BA2B": true,
			"number-9E19":  51,
			"number-FBEC":  31,
			"time-071A":    "2020-07-24 14:44:02",
			"text-DCD9":    "diyig1e",
		},
	}
	err := cli.SaveMany(dataMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
