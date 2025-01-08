package action

import (
	"reflect"
	"testing"
)

func TestNewRowFormSQLExecutor(t *testing.T) {
	type args struct {
		sqlTpl string
	}
	tests := []struct {
		name string
		args args
		want RowFormActionExecutor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRowFormSQLExecutor(tt.args.sqlTpl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRowFormSQLExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}
