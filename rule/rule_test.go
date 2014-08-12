package rule

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	r := &Rule{
		Ident:    "tmcnet",
		Start:    "http://www.tmcnet.com/voip/",
		CSSLinks: "a[href]",
		CSSTitle: "title",
		Restart:  30 * time.Minute,
		Accept: []*regexp.Regexp{
			regexp.MustCompile(`^/voip/(departments|columns|features)/articles/`),
		},
		Reject: []*regexp.Regexp{
			regexp.MustCompile(`bad link`),
		},
	}

	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		t.Fatalf("json.Marshal: %s", err)
	}
	t.Logf("JSON:\n%s", b)
}

func TestUnmarshal(t *testing.T) {
	data := `{
		"Ident": "moo",
		"Start": "http://moo.cow.net",
		"CSSLinks": "a[href]",
		"CSSTitle": "title",
		"RestartMins": 5,
		"Accept": [ "^/me", "^/me/too" ],
		"Reject": [ "reject me" ]
	}`

	r := new(Rule)
	if err := json.Unmarshal([]byte(data), r); err != nil {
		t.Fatal(err)
	}
	t.Logf("Rule:\n%+v", r)
}
