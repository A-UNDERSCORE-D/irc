package mode

import "strings"

// Type is one of any of the defined mode types. There *CAN* be others, which is why this is an int
// but I am nog really planning to test this.
type Type int

// Mode types
const (
	TypeA Type = iota
	TypeB
	TypeC
	TypeD
)

// ChannelMode represents a single channel mode, its type, and the prefix its displayed with, if applicable
type ChannelMode struct {
	Type   Type
	Char   string
	Prefix string
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
		out = append(out, makeModes(unknown, TypeD+Type(i+1))...)
	}

	return out
}

func makeModes(chars string, typ Type) (out []ChannelMode) {
	for _, c := range chars {
		out = append(out, ChannelMode{
			Type: typ,
			Char: string(c),
		})
	}

	return
}
