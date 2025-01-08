package idgen

import "testing"

func Test_sha256Generator_Generate(t *testing.T) {
	gen := NewSHA256Generator()
	param := []interface{}{"test123"}
	type args struct {
		param []interface{}
	}
	tests := []struct {
		name    string
		g       *sha256Generator
		args    args
		want    string
		wantErr bool
	}{
		{"default", gen, args{param: param}, "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("sha256Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sha256Generator.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
