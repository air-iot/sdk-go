package numberx

import (
	"reflect"
	"testing"
)

func TestGetValueByType(t *testing.T) {
	type args struct {
		valueType FieldType
		v         interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "float1",
			args: args{
				valueType: Float,
				v:         "1",
			},
			want:    float64(1),
			wantErr: false,
		},
		{
			name: "float2",
			args: args{
				valueType: Float,
				v:         "true",
			},
			want:    float64(1),
			wantErr: false,
		},
		{
			name: "int1",
			args: args{
				valueType: Int,
				v:         "1",
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueByType(tt.args.valueType, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValueByType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
