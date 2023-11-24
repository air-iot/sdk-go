package entity

import (
	"github.com/air-iot/json"
	"testing"
)

func Test_warn(t *testing.T) {
	//v.RegisterValidation("rfc3339", ValidateRFC3339)
	w1 := &WarnSend{ID: "1", Fields: []WarnTag{{Tag: Tag{ID: "test"}, Value: 1}}}
	//w1 := &Warn{ID: "1", Time: "2006-01-02T15:04:05Z07:00"}
	marshal, err := json.Marshal(w1)
	if err != nil {
		return
	}
	t.Log(string(marshal))
}
