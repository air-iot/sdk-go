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

func TestDataClient_GetLatest(t *testing.T) {
	cli := NewClient()
	r, err := cli.GetLatest([]map[string]interface{}{
		{"uid": "SDK1", "tagId": "SJD1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestDataClient_PostLatest(t *testing.T) {
	cli := NewClient()
	r, err := cli.PostLatest([]map[string]interface{}{
		{"uid": "SDK1", "tagId": "SJD1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestDataClient_GetQuery(t *testing.T) {
	cli := NewClient()
	r, err := cli.GetQuery([]map[string]interface{}{
		{"fields": []interface{}{"SJD1"}, "modelId": "5ea0fedee7fb6cf0e1907068", "where": []interface{}{"time > now()-1m"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestDataClient_PostQuery(t *testing.T) {
	cli := NewClient()
	r, err := cli.PostQuery([]map[string]interface{}{
		{"fields": []interface{}{"SJD1"}, "modelId": "5ea0fedee7fb6cf0e1907068", "where": []interface{}{"time > now()-1m"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}
