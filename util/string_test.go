package util

import (
	"reflect"
	"testing"
)

func TestChunkMessage(t *testing.T) {
	t.Parallel()

	type args struct {
		message   string
		maxLength int
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "one",
			args: args{
				message:   "this is a test",
				maxLength: 4,
			},
			want: []string{"this", " is ", "a te", "st"},
		},
		{
			name: "ret same",
			args: args{
				message:   "this is a test",
				maxLength: 1337,
			},
			want: []string{"this is a test"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ChunkMessage(tt.args.message, tt.args.maxLength); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChunkMessage() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
