package mode_test

import (
	"reflect"
	"testing"

	"awesome-dragon.science/go/irc/mode"
)

const (
	// Libera.chat's modes at time of writing, it doesnt really matter
	// what these are but its nice to have a full set anyway.

	aModes = "eIbq"
	bModes = "k"
	cModes = "flj"
	dModes = "CFLMPQScgimnprstuz"
)

func makeModes(chars string, typ mode.Type) (out mode.Set) {
	for _, c := range chars {
		out = append(out, mode.Mode{
			Type: typ,
			Char: c,
		})
	}

	return
}

func makeTestModes(t *testing.T) (out mode.Set) {
	t.Helper()

	out = append(out, makeModes(aModes, mode.TypeA)...)
	out = append(out, makeModes(bModes, mode.TypeB)...)
	out = append(out, makeModes(cModes, mode.TypeC)...)
	out = append(out, makeModes(dModes, mode.TypeD)...)
	// Just do prefixes manually
	out = append(
		out,
		mode.Mode{Type: mode.TypeD, Char: 'o', Prefix: "@"}, mode.Mode{Type: mode.TypeD, Char: 'v', Prefix: "+"},
	)

	return
}

func TestModeSet_ParseModeSequence(t *testing.T) { //nolint:funlen // Its a test
	t.Parallel()
	modeset := makeTestModes(t)
	tests := []struct {
		name string
		args string
		want mode.Sequence
	}{
		{
			name: "all adds type d",
			args: "+CFLgim",
			want: mode.Sequence{
				{Adding: true, Mode: modeset.GetMode('C')},
				{Adding: true, Mode: modeset.GetMode('F')},
				{Adding: true, Mode: modeset.GetMode('L')},
				{Adding: true, Mode: modeset.GetMode('g')},
				{Adding: true, Mode: modeset.GetMode('i')},
				{Adding: true, Mode: modeset.GetMode('m')},
			},
		},
		{
			name: "all adds type d trailing data",
			args: "+CFLgim thisShouldBeIgnored",
			want: mode.Sequence{
				{Adding: true, Mode: modeset.GetMode('C')},
				{Adding: true, Mode: modeset.GetMode('F')},
				{Adding: true, Mode: modeset.GetMode('L')},
				{Adding: true, Mode: modeset.GetMode('g')},
				{Adding: true, Mode: modeset.GetMode('i')},
				{Adding: true, Mode: modeset.GetMode('m')},
			},
		},
		{
			name: "z and q together",
			args: "+zq $~a",
			want: mode.Sequence{
				{Adding: true, Mode: modeset.GetMode('z')},
				{Adding: true, Mode: modeset.GetMode('q'), Parameter: "$~a"},
			},
		},
		{
			name: "bunch of various modes",
			args: "+CPcfntz #libera-overflow",
			want: mode.Sequence{
				{Adding: true, Mode: modeset.GetMode('C')},
				{Adding: true, Mode: modeset.GetMode('P')},
				{Adding: true, Mode: modeset.GetMode('c')},
				{Adding: true, Mode: modeset.GetMode('f'), Parameter: "#libera-overflow"},
				{Adding: true, Mode: modeset.GetMode('n')},
				{Adding: true, Mode: modeset.GetMode('t')},
				{Adding: true, Mode: modeset.GetMode('z')},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Because parallel
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := modeset.ParseModeSequence(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModeSet.ParseModeSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSequence_Collapse(t *testing.T) { //nolint:funlen // Its a test
	t.Parallel()
	modes := makeTestModes(t)
	tests := []struct {
		name string
		s    mode.Sequence
		want mode.Sequence
	}{
		{
			name: "all os",
			s:    modes.ParseModeSequence("+oooooooooooooooo"),
			want: modes.ParseModeSequence("+o"),
		},
		{
			name: "add and remove to end up with nothing",
			s:    modes.ParseModeSequence("+o-o"),
			want: modes.ParseModeSequence(""),
		},
		{
			name: "No changes",
			s:    modes.ParseModeSequence("+CPcfnt #libera-overflow"),
			want: modes.ParseModeSequence("+CPcfnt #libera-overflow"),
		},
		{
			name: "complex",
			s:    modes.ParseModeSequence("+CPcfntb #libera-overflow *!*@*"),
			want: modes.ParseModeSequence("+CPcfntb #libera-overflow *!*@*"),
		},
		{
			name: "complex dupe",
			s:    modes.ParseModeSequence("+CPcfntb-b #libera-overflow *!*@* *!*@*"),
			want: modes.ParseModeSequence("+CPcfnt #libera-overflow"),
		},
		{
			name: "multiple bans",
			s:    modes.ParseModeSequence("+bbbb a b c d"),
			want: modes.ParseModeSequence("+bbbb a b c d"),
		},
		{
			name: "multiple bans with dupes",
			s:    modes.ParseModeSequence("+bbbbbb a b c d a b"),
			want: modes.ParseModeSequence("+bbbb a b c d"),
		},
		{
			name: "multiple type as",
			s:    modes.ParseModeSequence("+qeb a b c"),
			want: modes.ParseModeSequence("+qeb a b c"),
		},
		{
			name: "multiple type to nothing",
			s:    modes.ParseModeSequence("+qeb-qeb a b c a b c"),
			want: modes.ParseModeSequence(""),
		},
		{
			name: "so I heard you like keys",
			s:    modes.ParseModeSequence("+kkkkkkkk a b c f g h i j"),
			want: modes.ParseModeSequence("+k j"),
		},
		{
			name: "so I heard you like keys, but it does nothing",
			s:    modes.ParseModeSequence("+kkkkkkkk-k a b c f g h i j j"),
			want: modes.ParseModeSequence(""),
		},
		{
			name: "snomask discard check",
			s:    modes.ParseModeSequence("+s +abcdefhij"),
			want: modes.ParseModeSequence("+s"),
		},
		{
			name: "beginning removal",
			s:    modes.ParseModeSequence("-C+C"),
			want: modes.ParseModeSequence("+C"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.s.Collapse(); !reflect.DeepEqual(got, tt.want) {
				t.Log("got  ", got)
				t.Log("want ", tt.want)
				t.Errorf("Sequence.Collapse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModesFromISupportToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want mode.Set
	}{
		{
			name: "one each",
			arg:  "a,b,c,d",
			want: mode.Set{
				mode.Mode{Type: mode.TypeA, Char: 'a'},
				mode.Mode{Type: mode.TypeB, Char: 'b'},
				mode.Mode{Type: mode.TypeC, Char: 'c'},
				mode.Mode{Type: mode.TypeD, Char: 'd'},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := mode.ModesFromISupportToken(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModesFromISupportToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
