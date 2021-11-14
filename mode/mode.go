package mode

import (
	"fmt"
	"strconv"
	"strings"

	"awesome-dragon.science/go/irc/util"
)

// Type is one of any of the defined mode types. There *CAN* be others, which is why this is an int
// but I am not really planning to test or support this.
type Type int

func (t Type) String() string {
	if t == -1 {
		return "Unknown"
	}

	if t <= TypeD {
		return modeTypeLUT[t]
	}

	return strconv.Itoa(int(t))
}

var modeTypeLUT = []string{TypeA: "A", TypeB: "B", TypeC: "C", TypeD: "D"} //nolint:gochecknoglobals // its a static LUT

// Mode types
const (
	TypeA Type = iota
	TypeB
	TypeC
	TypeD

	TypeUnknown Type = -1
)

// Mode represents a single channel mode, its type, and the prefix its displayed with, if applicable
type Mode struct {
	Type   Type
	Char   rune
	Prefix string
}

func (m Mode) String() string {
	prefix := ""
	if m.Prefix != "" {
		prefix = " P:" + m.Prefix
	}

	return fmt.Sprintf("{%s: T:%s%s}", string(m.Char), m.Type, prefix)
}

// Set is a []Mode with utility methods
type Set []Mode

func (m Set) String() string {
	s := make([]string, len(m))
	for i, v := range m {
		s[i] = v.String()
	}

	return fmt.Sprintf("[%s]", strings.Join(s, ","))
}

// GetMode returns the mode in the ModeSet represented by char, if it exists.
// if char does not exist as a mode, its type is returned as -1
func (m Set) GetMode(char rune) Mode {
	for _, mode := range m {
		if mode.Char == char {
			return mode
		}
	}

	return Mode{
		Type:   TypeUnknown,
		Char:   char,
		Prefix: "",
	}
}

// SequenceEntry is a list of mode changes
type SequenceEntry struct {
	Adding bool
	Mode
	Parameter string
}

func (s SequenceEntry) String() string {
	a := "+"
	if !s.Adding {
		a = "-"
	}

	param := ""
	if s.Parameter != "" {
		param = " " + strconv.Quote(s.Parameter)
	}

	return fmt.Sprintf("{%s %s%s}", a, s.Mode, param)
}

func popLeft(slice []string) (head string, rest []string) {
	if len(slice) == 0 {
		return "", slice
	}

	return slice[0], slice[1:]
}

func modeCount(s string) int {
	out := 0

	for r := range s {
		if r == '+' || r == '-' {
			continue
		}

		out++
	}

	return out
}

// Sequence is a sequence of mode changes
type Sequence []SequenceEntry

/*
from https://modern.ircdocs.horse/#mode-message
There are four categories of channel modes, defined as follows:

    Type A: Modes that add or remove an address to or from a list. These modes
		MUST always have a parameter when sent from the server to a client.
		A client MAY issue this type of mode without an argument to obtain the
		current contents of the list. The numerics used to retrieve contents of
		Type A modes depends on the specific mode. Also see the EXTBAN parameter.

	Type B: Modes that change a setting on a channel.
		These modes MUST always have a parameter.

	Type C: Modes that change a setting on a channel.
		These modes MUST have a parameter when being set,
		and MUST NOT have a parameter when being unset.

		Type D: Modes that change a setting on a channel.
		These modes MUST NOT have a parameter.

*/

// Collapse collapses a mode change Sequence into the eventual result of executing it
func (s Sequence) Collapse() Sequence { //nolint:funlen,gocognit,cyclop // Its just not possible to break this up more
	// This is a take 2, lets do this by applying ourselves to an empty sequence, and see how that goes
	seq := Sequence{}

	for _, op := range s {
		var toRemove []int

		found := false

		for i, other := range seq {
			if other.Char != op.Char {
				continue
			}

			found = true

			switch op.Type {
			case TypeA:
				// MUST have a parameter when sent server -> client, which is how we're assuming the string we got
				// was created. This is things like +b -- list modes, on libera thats `eIbq`
				//
				// as an aside, this also means that this can remove modes we dont think we know about, unless
				// we grab the entire list on join, but I dont think that's worth it for the core (and its definitely)
				// unrelated here
				if op.Parameter != other.Parameter {
					// they're unrelated, add and continue
					found = false

					continue
				}

				// okay, params are equal, if we're adding, and they are, ignore it, if we're adding and they're NOT,
				// remove it

			case TypeB:
				// MUST have a parameter, always, on libera this is just `k`, and for k specifically, on unset this is *
				// for this we're going to assume that unsets are correct, and sets just change it if its already set.
				// thus, this is a noop, as theres only *one* mode here, so its either added or removed, and as the
				// server is telling us this, we can believe it.
				// we want the last one to apply, so if both ops are adding, and the params dont match, update the param
				if other.Adding && op.Adding && other.Parameter != op.Parameter {
					seq[i].Parameter = op.Parameter // the last one is the one that matters if adding
				}

			case TypeC:
				// MUST have a parameter when set, MUST NOT when unset, but again, as this is from the server, we can be
				// lazy

			case TypeD:
				// never has a param, laziness intensifies!

			case TypeUnknown:
				// No idea! more lazy!

			default: // also no idea, but Im feeling exhaustive!
			}

			switch {
			case op.Adding && other.Adding:
				continue // we're both adding the exact same thing
			case !op.Adding && other.Adding:
				toRemove = append(toRemove, i) // we're removing whats already there, result is nothing changes

			case op.Adding && !other.Adding:
				seq[i].Adding = true
			}
		}

		if !found && op.Type == TypeA {
			// TypeA does this if its ignored; we want to be sure we dont add duplicates
			for _, v := range seq {
				if v.Char == op.Char && v.Parameter == op.Parameter {
					found = true
				}
			}
		}

		if !found {
			seq = append(seq, op)
		}

		// remove the requested indexes
		newSeq := make(Sequence, 0, len(seq))

		for i, v := range seq {
			if util.IntSliceContains(i, toRemove) {
				continue
			}

			newSeq = append(newSeq, v)
		}

		seq = newSeq
	}

	return seq
}

// ParseModeSequence parses a mode sequence to a useful set of mode changes
func (m Set) ParseModeSequence(sequence string) Sequence {
	var (
		modes string
		param string
	)

	adding := true
	split := strings.Split(sequence, " ")
	modes, split = popLeft(split)
	out := make(Sequence, 0, modeCount(modes))

	for _, r := range modes {
		if r == '+' || r == '-' {
			adding = r == '+'

			continue
		}

		param = ""

		mode := m.GetMode(r)
		switch mode.Type {
		case TypeUnknown, TypeD:
			// we dont know what this is, thus we assume it does *not* have
			// a parameter
			// or its TypeD, which also never has a parameter

		case TypeA, TypeB:
			// always has a parameter
			param, split = popLeft(split)

		case TypeC:
			// only has a parameter when setting
			if adding {
				param, split = popLeft(split)
			}
		}

		out = append(out, SequenceEntry{
			Adding:    adding,
			Mode:      mode,
			Parameter: param,
		})
	}

	return out
}

// ModesFromISupportToken creates a Mode array from an ISUPPORT MODE token
func ModesFromISupportToken(tokenArgs string) Set {
	var out []Mode

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

func makeModes(chars string, typ Type) (out []Mode) {
	for _, c := range chars {
		out = append(out, Mode{
			Type: typ,
			Char: c,
		})
	}

	return
}
