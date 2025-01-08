package idgen

import (
	"reflect"
	"testing"
)

func Test_uuidGeneratorV4_Generate(t *testing.T) {
	type args struct {
		param []interface{}
	}
	tests := []struct {
		name    string
		g       *uuidGeneratorV4
		args    args
		want    string
		wantErr bool
	}{
		{"default", NewUUIDGeneratorV4(), args{nil}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("uuidGeneratorV4.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("uuidGeneratorV4.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUUIDGeneratorV4(t *testing.T) {
	tests := []struct {
		name string
		want *uuidGeneratorV4
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUUIDGeneratorV4(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUUIDGeneratorV4() = %v, want %v", got, tt.want)
			}
		})
	}
}
