package idgen

import (
	"reflect"
	"testing"
)

func TestNewUUIDGeneratorV1(t *testing.T) {
	tests := []struct {
		name string
		want *uuidGeneratorV1
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUUIDGeneratorV1(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUUIDGeneratorV1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_uuidGeneratorV1_Generate(t *testing.T) {
	type args struct {
		param []interface{}
	}
	tests := []struct {
		name    string
		g       *uuidGeneratorV1
		args    args
		want    string
		wantErr bool
	}{
		{"defalt", NewUUIDGeneratorV1(), args{nil}, "", false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("uuidGeneratorV1.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("uuidGeneratorV1.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
