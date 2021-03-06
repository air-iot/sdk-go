package driver

import (
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
	var FixedS = "boundary"

	point := Point{
		Uid:     "b1",
		ModelId: "5fb887123ab8daa563692579",
		NodeId:  "5fb8874b3ab8daa56369257d",
		Fields: []Field{
			{Tag: Tag{
				ID:       "p1",
				Name:     "p1",
				TagValue: &TagValue{
					//MinValue: &minValue,
					//MaxValue: &MaxValue,
					//MinRaw:   &MinRaw,
					//MaxRaw:   &MaxRaw,
				},
				Fixed: &Fixed,
				Mod:   &Mod,
				Range: &Range{
					MinValue:   &minValue,
					MaxValue:   &MaxValue,
					Active:     &FixedS,
					FixedValue: &Mod,
				},
			},
				Value: 10.121,
			},
		},
	}

	app := NewApp()
	err := app.WritePoints(point)
	if err != nil {
		t.Fatal(err)
	}
}
