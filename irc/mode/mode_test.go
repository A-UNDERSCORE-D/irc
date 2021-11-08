package mode

import (
	"reflect"
	"testing"
)

const (
	// Libera.chat's modes at time of writing, it doesnt really matter
	// what these are but its nice to have a full set anyway.

	aModes = "eIbq"
	bModes = "k"
	cModes = "flj"
	dModes = "CFLMPQScgimnprstuz"
)

func makeTestModes(t *testing.T) (out ModeSet) {
	t.Helper()

	out = append(out, makeModes(aModes, TypeA)...)
	out = append(out, makeModes(bModes, TypeB)...)
	out = append(out, makeModes(cModes, TypeC)...)
	out = append(out, makeModes(dModes, TypeD)...)
	// Just do prefixes manually
	out = append(out, Mode{Type: TypeD, Char: 'o', Prefix: "@"}, Mode{Type: TypeD, Char: 'v', Prefix: "+"})

	return
}

func TestModeSet_ParseModeSequence(t *testing.T) { //nolint:funlen // Its a test
	t.Parallel()
	modeset := makeTestModes(t)
	tests := []struct {
		name string
		args string
		want []SequenceEntry
	}{
		{
			name: "all adds type d",
			args: "+CFLgim",
			want: []SequenceEntry{
				{adding: true, Mode: modeset.GetMode('C')},
				{adding: true, Mode: modeset.GetMode('F')},
				{adding: true, Mode: modeset.GetMode('L')},
				{adding: true, Mode: modeset.GetMode('g')},
				{adding: true, Mode: modeset.GetMode('i')},
				{adding: true, Mode: modeset.GetMode('m')},
			},
		},
		{
			name: "all adds type d trailing data",
			args: "+CFLgim thisShouldBeIgnored",
			want: []SequenceEntry{
				{adding: true, Mode: modeset.GetMode('C')},
				{adding: true, Mode: modeset.GetMode('F')},
				{adding: true, Mode: modeset.GetMode('L')},
				{adding: true, Mode: modeset.GetMode('g')},
				{adding: true, Mode: modeset.GetMode('i')},
				{adding: true, Mode: modeset.GetMode('m')},
			},
		},
		{
			name: "z and q together",
			args: "+zq $~a",
			want: []SequenceEntry{
				{adding: true, Mode: modeset.GetMode('z')},
				{adding: true, Mode: modeset.GetMode('q'), Parameter: "$~a"},
			},
		},
		{
			name: "bunch of various modes",
			args: "+CPcfntz #libera-overflow",
			want: []SequenceEntry{
				{adding: true, Mode: modeset.GetMode('C')},
				{adding: true, Mode: modeset.GetMode('P')},
				{adding: true, Mode: modeset.GetMode('c')},
				{adding: true, Mode: modeset.GetMode('f'), Parameter: "#libera-overflow"},
				{adding: true, Mode: modeset.GetMode('n')},
				{adding: true, Mode: modeset.GetMode('t')},
				{adding: true, Mode: modeset.GetMode('z')},
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
		s    Sequence
		want Sequence
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
