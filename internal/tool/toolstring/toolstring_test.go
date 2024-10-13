// Package toolstring string tool

package toolstring

import "testing"

func TestIsValidId(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.

		{"test1", args{"qwe"}, true},
		{"test2", args{"-"}, false},
		{"test3", args{"_"}, true},
		{"test4", args{""}, false},
		{"test5", args{"qwe/../"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidID(tt.args.value); got != tt.want {
				t.Errorf("IsVarName() = %v, want %v", got, tt.want)
			}
		})
	}
}
