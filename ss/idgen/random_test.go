package idgen

import "testing"

func Test_randomIDGenerator_Generate(t *testing.T) {
	type args struct {
		param []interface{}
	}
	tests := []struct {
		name    string
		g       *randomIDGenerator
		args    args
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("randomIDGenerator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("randomIDGenerator.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
