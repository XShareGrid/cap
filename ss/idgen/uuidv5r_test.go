package idgen

import "testing"

func Test_uuidGeneratorV5Recurrent_Generate(t *testing.T) {
	type args struct {
		param []interface{}
	}
	params := []interface{}{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", "test", "1", "2"}
	tests := []struct {
		name    string
		g       *uuidGeneratorV5Recurrent
		args    args
		want    string
		wantErr bool
	}{

		{"default", NewUUIDGeneratorV5Recurrent(), args{params}, "0e4163da-6cdd-56b2-8bb5-cf4817ef209c", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.Generate(tt.args.param...)
			if (err != nil) != tt.wantErr {
				t.Errorf("uuidGeneratorV5.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("uuidGeneratorV5.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
