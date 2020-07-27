package api

import (
	"testing"

	"github.com/air-iot/sdk-go/traefik"
)

func init() {
	traefik.Host = "iot.tmis.top"
	traefik.Port = 8010
	traefik.Enable = true
	traefik.AppKey = "b9bd592b-2d79-4f5c-d583-aad18ebe00ca"
	traefik.AppSecret = "c5de1068-79fd-b32b-a4f8-291c337111fa"
}

func TestExtClient_FindById(t *testing.T) {
	cli := NewClient()
	var r = make(map[string]interface{})
	err := cli.FindExtById("新表", "5f1aa80eac624e29f1678fd3", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestExtClient_SaveMany(t *testing.T) {
	cli := NewClient()
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
	err := cli.SaveManyExt("新表", dataMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
