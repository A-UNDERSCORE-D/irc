package mode

import "strings"

type ModeType int

const (
	TypeA ModeType = iota
	TypeB
	TypeC
	TypeD
)

type ChannelMode struct {
	Type ModeType
	Char string
}

// ModesFromISupportToken creates a ChannelMode array from an ISUPPORT MODE token
func ModesFromISupportToken(tokenArgs string) []ChannelMode {
	var out []ChannelMode

	split := strings.Split(tokenArgs, ",")
	A := split[0]
	B := split[1]
	C := split[2]
	D := split[3]
	other := []string{}

	if len(split) > 4 {
		other = split[4:]
	}

	out = append(out, makeModes(A, TypeA)...)
	out = append(out, makeModes(B, TypeB)...)
	out = append(out, makeModes(C, TypeC)...)
	out = append(out, makeModes(D, TypeD)...)

	for i, unknown := range other {
		out = append(out, makeModes(unknown, TypeD+ModeType(i+1))...)
	}

	return out
}

func makeModes(chars string, typ ModeType) (out []ChannelMode) {
	for _, c := range chars {
		out = append(out, ChannelMode{
			Type: typ,
			Char: string(c),
		})
	}

	return
}
