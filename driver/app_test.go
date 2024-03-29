package driver

import (
	"context"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"github.com/spf13/viper"
	"testing"
)

func TestApp_WritePoints(t *testing.T) {
	viper.Set("env", "dev")
	viper.Set("mqtt.host", "127.0.0.1")
	viper.Set("driver.id", "test")
	viper.Set("driver.name", "test")
	var minValue float64 = 10
	var MaxValue float64 = 100
	//var MinRaw float64 = 200
	//var MaxRaw float64 = 300
	var Fixed int32 = 2
	var Mod float64 = 2
	point := entity.Point{
		ID: "b1",
		Fields: []entity.Field{
			{Tag: entity.Tag{
				ID:       "p1",
				Name:     "p1",
				TagValue: &entity.TagValue{
					//MinValue: &minValue,
					//MaxValue: &MaxValue,
					//MinRaw:   &MinRaw,
					//MaxRaw:   &MaxRaw,
				},
				Fixed: &Fixed,
				Mod:   &Mod,
				Range: &entity.Range{
					MinValue:   &minValue,
					MaxValue:   &MaxValue,
					Active:     entity.Active_Boundary,
					FixedValue: &Mod,
				},
			},
				Value: 10.121,
			},
		},
	}

	a := NewApp()
	err := a.WritePoints(context.Background(), point)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wp(t *testing.T) {
	a := new(app)
	var minValue float64 = 10
	var MaxValue float64 = 100
	//var MinRaw float64 = 200
	//var MaxRaw float64 = 300
	var Fixed int32 = 2
	var Mod float64 = 2
	point := entity.Point{
		ID: "b1",
		Fields: []entity.Field{
			{Tag: entity.Tag{
				ID:       "p1",
				Name:     "p1",
				TagValue: &entity.TagValue{
					//MinValue: &minValue,
					//MaxValue: &MaxValue,
					//MinRaw:   &MinRaw,
					//MaxRaw:   &MaxRaw,
				},
				Fixed: &Fixed,
				Mod:   &Mod,
				Range: &entity.Range{
					MinValue:   &minValue,
					MaxValue:   &MaxValue,
					Active:     entity.Active_Boundary,
					FixedValue: &Mod,
				},
			},
				Value: true,
			},
		},
	}
	err := a.writePoints(context.Background(), "test", point)
	if err != nil {
		t.Fatal(err)
	}
}
