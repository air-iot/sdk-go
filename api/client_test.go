/**
 * @Author: ZhangQiang
 * @Description:
 * @File:  client_test
 * @Version: 1.0.0
 * @Date: 2020/8/4 16:49
 */
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

var cli = NewClient("http", "localhost", 31001, "60bf02235196c2b9ae932f01", "e2d28203-1fdf-8442-a380-31cff9c1ac4a", "33be8f2f-5df8-a3b3-d05a-329f269628d3")

func TestClient_GetLatest(t *testing.T) {
	r, err := cli.GetLatest([]map[string]interface{}{
		{"uid": "SDK1", "tagId": "SJD1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestClient_PostLatest(t *testing.T) {
	r, err := cli.PostLatest([]map[string]interface{}{
		{"uid": "SDK1", "tagId": "SJD1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestClient_GetQuery(t *testing.T) {
	r, err := cli.GetQuery([]map[string]interface{}{
		{"fields": []interface{}{"SJD1"}, "modelId": "5ea0fedee7fb6cf0e1907068", "where": []interface{}{"time > now()-1m"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestClient_PostQuery(t *testing.T) {
	r, err := cli.PostQuery([]map[string]interface{}{
		{"fields": []interface{}{"SJD1"}, "modelId": "5ea0fedee7fb6cf0e1907068", "where": []interface{}{"time > now()-1m"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestEvent_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindEventQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestEvent_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindEventById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestEvent_Save(t *testing.T) {

	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveEvent(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestEvent_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceEventById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestEvent_UpdateById(t *testing.T) {

	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateEventById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestEvent_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelEventById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestExtClient_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindExtById("新表", "5f1aa80eac624e29f1678fd3", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestExtClient_SaveMany(t *testing.T) {
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

func TestHandler_FindQuery(t *testing.T) {
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
	var r = make(map[string]interface{})
	err := cli.FindHandlerById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestHandler_Save(t *testing.T) {
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
	var r = make(map[string]interface{})
	err := cli.DelHandlerById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindModelQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindModelById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_Save(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveModel(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceModelById("5f1aae1eac624e29f1678fd5", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_UpdateById(t *testing.T) {

	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateModelById("5f1aae1eac624e29f1678fd5", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestModelClient_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelModelById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindNodeQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindNodeById("001", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_Save(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveNode(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceNodeById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_UpdateById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateNodeById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestNode_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelNodeById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindSettingQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindSettingById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_Save(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveSetting(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceSettingById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_UpdateById(t *testing.T) {

	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateSettingById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestSetting_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelSettingById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindTableQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindTableById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_Save(t *testing.T) {

	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveTable(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceTableById("5f1aae1eac624e29f1678fd5", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_UpdateById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateTableById("5f1aae1eac624e29f1678fd5", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestTable_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelTableById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_FindQuery(t *testing.T) {
	var r = make([]map[string]interface{}, 0)
	query := `{"filter":{"device.driver":"test","$lookups":[{"from":"node","localField":"_id","foreignField":"model","as":"devices"},{"from":"node","localField":"devices.parent","foreignField":"_id","as":"devicesParent"},{"from":"model","localField":"devicesParent.model","foreignField":"_id","as":"devicesParentModel"}]},"project":{"device":1,"devices":1,"devicesParent":1,"devicesParentModel":1}}`

	queryMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(query), &queryMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.FindUserQuery(&queryMap, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_FindById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.FindUserById("5ecf1f423e951ef12218381d", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_Save(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK1","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.SaveUser(data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_ReplaceById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK2","kind":"7c6b7a0f-998e-445d-ab00-fd9cd69ee051","device":{"driver":"test","settings":{"interval":3,"network":{}},"tags":[{"name":"数据点A","rules":{"high":1},"unit":"m","id":"SJD1"},{"id":"SJD2","name":"数据点B","unit":"c"},{"id":"SJD3","name":"数据C"},{"id":"SJD4","name":"数据点D"},{"id":"SJD5","name":"数据点E"},{"id":"SJD6","name":"数据点F"},{"name":"数据点G","id":"SJD7"},{"id":"SJD8","name":"数据点H"},{"id":"SJD9","name":"数据点I"},{"name":"数据点J","id":"SJD10"}]},"type":[],"computed":{"tags":[]},"order":212,"statusList":[{"focus":false,"user":"5c74edbc6f553e4fca5df9c6"}],"table":{"colors":{"timeout":"#8e7cc3","offline":"#e69138","warning1":"#4c2f0a","warning2":"#ff6347","warning3":"#f00","bg":"transparent","normal":"#000"},"fields":[{"key":"uid","title":"编号"},{"key":"param-SJD1","title":"数据点A"},{"key":"param-SJD2","title":"数据点B"}]}}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ReplaceUserById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_UpdateById(t *testing.T) {
	var r = make(map[string]interface{})

	data := `{"name":"SDK3"}`
	dataMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UpdateUserById("5ece2b44e1fe4ebf858a778c", data, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUser_DelById(t *testing.T) {
	var r = make(map[string]interface{})
	err := cli.DelUserById("5ece2b44e1fe4ebf858a778c", &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestClient_UploadMediaFile(t *testing.T) {
	createFile := func() (string, error) {
		filename := fmt.Sprintf("%s%cmedia_upload_file.txt", os.TempDir(), os.PathSeparator)
		tmpFile, err := os.Create(filename)
		if err != nil {
			return "", err
		}

		if _, err = tmpFile.WriteString("the file is used to test upload of media library"); err != nil {
			return "", err
		}

		if err = tmpFile.Close(); err != nil {
			return "", err
		}

		return tmpFile.Name(), nil
	}

	filename, err := createFile()
	if err != nil {
		t.Fatalf("cannot create tmp file, %+v", err)
	}

	t.Logf("create a tmp file: %s", filename)

	defer func() {
		_ = os.Remove(filename)
	}()

	sep := bytes.LastIndexByte([]byte(filename), os.PathSeparator)
	name := filename[sep+1:]

	tmpFile, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Fatalf("open tmp file %s failed, %+v", filename, err)
	}

	defer tmpFile.Close()

	var client = NewClient("http", "airiot.tech", 31001, "62c4f8aaa0f974f96cca7ddc", "2963577e-d0bb-c8ba-ed5e-7574cd78fbf7", "f160a24b-1780-89e7-cd48-b4c0073bd0fe")
	url, err := client.UploadMediaFile(name, "chenpc", "cover", tmpFile)
	if err != nil {
		t.Fatalf("upload file failed, %+v", err)
	}

	if len(url) == 0 {
		t.Fatalf("the url of uploaded file is empty")
	}

	t.Logf("url: %s", url)
}

func TestClient_GetMediaFile(t *testing.T) {
	url := "/core/fileServer/mediaLibrary/62c4f8aaa0f974f96cca7ddc/chenpc/media_upload_file.txt"
	var client = NewClient("http", "airiot.tech", 31001, "62c4f8aaa0f974f96cca7ddc", "2963577e-d0bb-c8ba-ed5e-7574cd78fbf7", "f160a24b-1780-89e7-cd48-b4c0073bd0fe")
	file, err := client.GetMediaFile(url)
	if err != nil {
		t.Fatalf("download file %s failed, %+v", url, err)
	}

	if strings.Compare(file.Name, "media_upload_file.txt") != 0 {
		t.Fatalf("the filename mismatch, expected: media_upload_file.txt, got: %s", file.Name)
	}

	t.Logf("file size: %d", file.Size)

	defer func() {
		file.Close()
	}()

	buf, err := file.ReadAll()
	if err != nil {
		t.Fatalf("cannot read file, %+v", err)
	}

	var size = int64(len(buf))
	t.Logf("read size: %d", size)

	if file.Size != 0 && file.Size != size {
		t.Fatalf("invalid file size, expected: %d, got: %d", file.Size, size)
	}

	content := string(buf[:size])

	if strings.Compare(content, "the file is used to test upload of media library") != 0 {
		t.Fatalf("the content of file mismatch, expected: 'the file is used to test upload of media library', got: '%s'", content)
	}
}

func TestClient_DeleteMediaFile(t *testing.T) {
	path := "/chenpc/media_upload_file.txt"
	var client = NewClient("http", "airiot.tech", 31001, "62c4f8aaa0f974f96cca7ddc", "2963577e-d0bb-c8ba-ed5e-7574cd78fbf7", "f160a24b-1780-89e7-cd48-b4c0073bd0fe")
	err := client.DeleteMediaFile(path, true)
	if err != nil {
		t.Fatalf("delete file %s failed, %+v", path, err)
	}
}
