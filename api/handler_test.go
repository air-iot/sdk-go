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

func TestHandler_FindQuery(t *testing.T) {
	cli := NewClient()
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindHandlerQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_FindById(t *testing.T) {
	cli := NewClient()
	var r = make(map[string]interface{})
	err := cli.FindHandlerById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_Save(t *testing.T) {
	cli := NewClient()

	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveHandler(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_ReplaceById(t *testing.T) {
	cli := NewClient()
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceHandlerById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_UpdateById(t *testing.T) {
	cli := NewClient()

	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateHandlerById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_DelById(t *testing.T) {
	cli := NewClient()
	var r = make(map[string]interface{})
	err := cli.DelHandlerById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
