package chatcommand_test

import (
	"testing"

	"awesome-dragon.science/go/irc/event/chatcommand"
)

func TestArgument_ArgString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a    *chatcommand.Argument
		want string
	}{
		{
			name: "nil",
			a:    &chatcommand.Argument{Arguments: nil},
			want: "",
		},
		{
			name: "empty",
			a:    &chatcommand.Argument{Arguments: []string{}},
			want: "",
		},
		{
			name: "single",
			a:    &chatcommand.Argument{Arguments: []string{"test"}},
			want: "test",
		},
		{
			name: "many",
			a: &chatcommand.Argument{
				Arguments: []string{"it", "was", "the", "dawn", "of", "the", "third", "age", "of", "mankind"},
			},
			want: "it was the dawn of the third age of mankind",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.ArgString(); got != tt.want {
				t.Errorf("Argument.ArgString() = %v, want %v", got, tt.want)
			}
		})
	}
}
