package isupport_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"awesome-dragon.science/go/irc/isupport"
	"github.com/ergochat/irc-go/ircmsg"
)

// for Mode, wherever that is
// wantModes := mode.ModeSet{
// 	mode.Mode{Type: mode.TypeA, Char: 'a'},
// 	mode.Mode{Type: mode.TypeB, Char: 'b'},
// 	mode.Mode{Type: mode.TypeC, Char: 'c'},
// 	mode.Mode{Type: mode.TypeD, Char: 'd'},
// 	mode.Mode{Type: mode.TypeD, Char: 'o', Prefix: '@'},
// }

// is := isupport.New()
// is.Parse(parseLineMust(":a.b.c 005 test CHANMODES=a,b,c,d PREFIX=(@)o"))

// res := is.ChanModes()
// for i := 0; i < len(wantModes); i++ {
// 	if res[i] != wantModes[i] {

// 	}
// }

var iSupport *isupport.ISupport //nolint:gochecknoglobals // its the same for every test, for the most part

func parseLineMust(raw string) *ircmsg.Message {
	line, err := ircmsg.ParseLine(raw)
	if err != nil {
		panic(err)
	}

	return &line
}

func makeIS(tokens ...string) *isupport.ISupport {
	out := isupport.New()
	for _, t := range tokens {
		out.Parse(&ircmsg.Message{
			Source:  ":dragon.libera.chat",
			Command: "005",
			Params:  []string{"test", t, "are supported by this server"},
		})
	}

	return out
}

func TestMain(m *testing.M) {
	iSupport = isupport.New()

	//nolint:lll // they have to be
	lines := []*ircmsg.Message{
		parseLineMust(":dragon.libera.chat 005 TEST FNC KNOCK CALLERID=g SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE CHANTYPES=# EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLMPQScgimnprstuz :are supported by this server"),
		parseLineMust(":dragon.libera.chat 005 TEST CHANLIMIT=#:250 PREFIX=(ov)@+ MAXLIST=bqeI:100 MODES=4 NETWORK=Libera.Chat STATUSMSG=@+ CASEMAPPING=rfc1459 NICKLEN=16 MAXNICKLEN=16 CHANNELLEN=50 TOPICLEN=390 DEAF=D :are supported by this server"),
		parseLineMust(":dragon.libera.chat 005 TEST TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,ajrxz :are supported by this server"),
		// Ones that werent otherwise set, that we want for testing
		parseLineMust(":dragon.libera.chat 005 TEST AWAYLEN=1337 :are supported by this server"),
	}

	for _, l := range lines {
		iSupport.Parse(l)
	}

	os.Exit(m.Run())
}

func TestISupport_MaxAwayLen(t *testing.T) {
	t.Parallel()

	const want = 1337
	if res := iSupport.MaxAwayLen(); res != want {
		t.Errorf("ISupport.MaxAwayLen() = %d, want %d", res, want)
	}
}

func TestISupport_CaseMapping(t *testing.T) {
	t.Parallel()

	const want = "rfc1459"
	if res := iSupport.CaseMapping(); res != want {
		t.Errorf("ISupport.CaseMapping() = %q, want %q", res, want)
	}
}

func TestISupport_ChanLimit(t *testing.T) {
	t.Parallel()

	want := map[string]int{"#": 250}
	if res := iSupport.ChanLimit(); fmt.Sprint(res) != fmt.Sprint(want) {
		t.Errorf("ISupport.ChanLimit() = %v, want %v", res, want)
	}
}

func TestISupport_ChanModes(t *testing.T) {
	t.Parallel()

	const want = "eIbq,k,flj,CFLMPQScgimnprstuz"
	if res := iSupport.ChanModes(); res != want {
		t.Errorf("ISupport.ChanModes() = %s, want %s", res, want)
	}
}

func TestISupport_MaxChanLen(t *testing.T) {
	t.Parallel()

	const want = 50
	if res := iSupport.MaxChanLen(); res != want {
		t.Errorf("ISupport.MaxChanLen() = %d, want %d", res, want)
	}
}

func TestISupport_ChanTypes(t *testing.T) {
	t.Parallel()

	if want, res := []string{"#"}, iSupport.ChanTypes(); !reflect.DeepEqual(want, res) {
		t.Errorf("ISupport.ChanTypes() = %#v, want %#v", res, want)
	}
}

func TestISupport_EList(t *testing.T) {
	t.Parallel()

	if want, res := []string{"C", "T", "U"}, iSupport.EList(); !reflect.DeepEqual(want, res) {
		t.Errorf("iSupport.EList() = %#v, want %#v", res, want)
	}
}

func TestISupport_Excepts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want string
	}{
		{name: "LC", is: iSupport, want: "e"},
		{name: "default", is: makeIS("EXCEPTS"), want: "e"},
		{name: "none", is: isupport.New(), want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.Excepts(); got != tt.want {
				t.Errorf("ISupport.Excepts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_Extban(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		is         *isupport.ISupport
		wantPrefix string
		wantTypes  []string
	}{
		{
			name:       "libera",
			is:         iSupport,
			wantPrefix: "$",
			wantTypes:  []string{"a", "j", "r", "x", "z"},
		},
		{
			name:       "no prefix",
			is:         makeIS("EXTBAN=,a"),
			wantPrefix: "",
			wantTypes:  []string{"a"},
		},
		{
			name:       "none",
			is:         makeIS(),
			wantPrefix: "",
			wantTypes:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPrefix, gotTypes := tt.is.Extban()
			if gotPrefix != tt.wantPrefix {
				t.Errorf("ISupport.Extban() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if !reflect.DeepEqual(gotTypes, tt.wantTypes) {
				t.Errorf("ISupport.Extban() gotTypes = %v, want %v", gotTypes, tt.wantTypes)
			}
		})
	}
}

func TestISupport_MaxHostLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: -1},
		{name: "libera", is: makeIS("HOSTLEN=10"), want: 10},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxHostLen(); got != tt.want {
				t.Errorf("ISupport.MaxHostLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_InviteExemption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want string
	}{
		{name: "libera", is: iSupport, want: "I"},
		{name: "nondefault", is: makeIS("INVEX=E"), want: "E"},
		{name: "default", is: makeIS("INVEX"), want: "I"},
		{name: "none", is: makeIS(), want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.InviteExemption(); got != tt.want {
				t.Errorf("ISupport.InviteExemption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxKickLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: -1},
		{name: "exists", is: makeIS("KICKLEN=10"), want: 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxKickLen(); got != tt.want {
				t.Errorf("ISupport.MaxKickLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxListModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want map[rune]int
	}{
		{name: "libera", is: iSupport, want: map[rune]int{'b': 100, 'q': 100, 'e': 100, 'I': 100}},
		{name: "none", is: makeIS(), want: nil},
		{name: "multiple", is: makeIS("MAXLIST=x:10,v:5"), want: map[rune]int{'x': 10, 'v': 5}},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxListModes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ISupport.MaxListModes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: -1},
		{name: "set", is: makeIS("MAXTARGETS=10"), want: 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxTargets(); got != tt.want {
				t.Errorf("ISupport.MaxTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: 4},
		{name: "unset", is: makeIS(), want: -1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxModes(); got != tt.want {
				t.Errorf("ISupport.MaxModes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_Network(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want string
	}{
		{name: "libera", is: iSupport, want: "Libera.Chat"},
		{name: "unset", is: makeIS(), want: ""},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.Network(); got != tt.want {
				t.Errorf("ISupport.Network() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxNickLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: 16},
		{name: "unset", is: makeIS(), want: -1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.MaxNickLen(); got != tt.want {
				t.Errorf("ISupport.MaxNickLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_Prefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want map[rune]rune
	}{
		{name: "libera", is: iSupport, want: map[rune]rune{'o': '@', 'v': '+'}},
		{name: "none", is: makeIS(), want: nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.Prefix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ISupport.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_SafeList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want bool
	}{
		{name: "libera", is: iSupport, want: true},
		{name: "exists", is: makeIS("SAFELIST"), want: true},
		{name: "noexist", is: makeIS(), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.SafeList(); got != tt.want {
				t.Errorf("ISupport.SafeList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_SilenceMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: -1},
		{name: "libera", is: makeIS("SILENCE=1337"), want: 1337},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.SilenceMax(); got != tt.want {
				t.Errorf("ISupport.SilenceMax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_StatusMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want []string
	}{
		{name: "libera", is: iSupport, want: []string{"@", "+"}},
		{name: "none", is: makeIS(), want: []string{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.StatusMsg(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ISupport.StatusMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxCommandTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want map[string]int
	}{
		{
			name: "libera",
			is:   iSupport,
			want: map[string]int{
				"names": 1, "list": 1, "kick": 1, "whois": 1, "privmsg": 4,
				"notice": 4, "accept": -1, "monitor": -1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxCommandTargets(); !reflect.DeepEqual(got, tt.want) {
				t.Log(got)
				t.Log(tt.want)
				t.Errorf("ISupport.MaxCommandTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestISupport_MaxTopicLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		is   *isupport.ISupport
		want int
	}{
		{name: "libera", is: iSupport, want: 390},
		{name: "unset", is: makeIS(), want: -1},
		{name: "random", is: makeIS("TOPICLEN=1337"), want: 1337},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.is.MaxTopicLen(); got != tt.want {
				t.Errorf("ISupport.MaxTopicLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
