package idgen

import "testing"

func Test_md5Generator_Generate(t *testing.T) {
	gen := NewMD5Generator()
	param := []interface{}{"test123"}
	type args struct {
		param []interface{}
	}
	tests := []struct {
		name    string
		g       *md5Generator
		args    args
		want    string
		wantErr bool
	}{
		{"default", gen, args{param: param}, "cc03e747a6afbbcbf8be7668acfebee5", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("md5Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("md5Generator.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
