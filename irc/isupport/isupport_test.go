// Package isupport contains an implementation of an ISUPPORT message handler
package isupport_test

import (
	"reflect"
	"testing"

	"awesome-dragon.science/go/irc/irc/isupport"
	"awesome-dragon.science/go/irc/irc/mode"
	"github.com/ergochat/irc-go/ircmsg"
)

func getIsupport(t *testing.T) *isupport.ISupport {
	t.Helper()

	out := isupport.New()

	//nolint:lll // they have to be
	lines := []*ircmsg.Message{
		parseLineMust(":dragon.libera.chat 005 TEST FNC KNOCK CALLERID=g SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE CHANTYPES=# EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLMPQScgimnprstuz :are supported by this server"),
		parseLineMust(":dragon.libera.chat 005 TEST CHANLIMIT=#:250 PREFIX=(ov)@+ MAXLIST=bqeI:100 MODES=4 NETWORK=Libera.Chat STATUSMSG=@+ CASEMAPPING=rfc1459 NICKLEN=16 MAXNICKLEN=16 CHANNELLEN=50 TOPICLEN=390 DEAF=D :are supported by this server"),
		parseLineMust(":dragon.libera.chat 005 TEST TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,ajrxz :are supported by this server"),
		// Ones that werent otherwise set, that we want for testing
		parseLineMust(":dragon.libera.chat 005 TEST AWAYLEN=1337 :are supported by this server"),
	}

	for _, l := range lines {
		out.Parse(l)
	}

	return out
}

func TestNew(t *testing.T) {
	t.Parallel()
	isupport.New()
}

func TestISupport_Parse(t *testing.T) {
	t.Parallel()

	x := isupport.New()
	x.Parse(parseLineMust(":test 005 A TEST :are supported on this server"))

	if !x.HasToken("TEST") {
		t.Error("Expected TEST to exist")
	}
}

func TestISupport_Modes(t *testing.T) {
	t.Parallel()

	wantModes := mode.ModeSet{
		mode.Mode{Type: mode.TypeA, Char: 'a'},
		mode.Mode{Type: mode.TypeB, Char: 'b'},
		mode.Mode{Type: mode.TypeC, Char: 'c'},
		mode.Mode{Type: mode.TypeD, Char: 'd'},
		mode.Mode{Type: mode.TypeD, Char: 'o', Prefix: "@"},
	}

	is := isupport.New()
	is.Parse(parseLineMust(":a.b.c 005 test CHANMODES=a,b,c,d PREFIX=(@)o :are supported by this server"))

	if res := is.Modes(); !reflect.DeepEqual(res, wantModes) {
		t.Errorf("is.Modes() = %v, want %v", res, wantModes)
	}
}

func TestISupport_GetToken(t *testing.T) {
	t.Parallel()

	is := getIsupport(t)
	tests := []struct {
		name       string
		is         *isupport.ISupport
		arg        string
		wantValue  string
		wantExists bool
	}{
		{
			name:       "libera",
			is:         is,
			arg:        "INVEX",
			wantValue:  "",
			wantExists: true,
		},
		{
			name:       "libera-noexist",
			is:         is,
			arg:        "doesnt-exist",
			wantValue:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotValue, gotExists := tt.is.GetToken(tt.arg)
			if gotValue != tt.wantValue {
				t.Errorf("ISupport.GetToken() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotExists != tt.wantExists {
				t.Errorf("ISupport.GetToken() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestISupport_GetTokenDefault(t *testing.T) {
	t.Parallel()

	is := getIsupport(t)
	tests := []struct {
		name      string
		is        *isupport.ISupport
		tokenName string
		dflt      string
		want      string
	}{
		{
			name:      "exists",
			is:        is,
			tokenName: "INVEX",
			dflt:      "A",
			want:      "A",
		},
		{
			name:      "no exist",
			is:        is,
			tokenName: "NOEXIST",
			dflt:      "Asd",
			want:      "",
		},
		{
			name:      "no exist",
			is:        is,
			tokenName: "CHANTYPES",
			dflt:      "&",
			want:      "#",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.is.GetTokenDefault(tt.tokenName, tt.dflt); got != tt.want {
				t.Errorf("ISupport.GetTokenDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
