package purell

import (
	"testing"
)

// Test cases merged from PR #1
// Originally from https://github.com/jehiah/urlnorm/blob/master/test_urlnorm.py

func assertMap(t *testing.T, cases map[string]string, f NormalizationFlags) {
	for bad, good := range cases {
		s, e := NormalizeURLString(bad, f)
		if e != nil {
			t.Errorf("%s normalizing %v to %v", e.Error(), bad, good)
		} else {
			if s != good {
				t.Errorf("source: %v expected: %v got: %v", bad, good, s)
			}
		}
	}
}

// This tests normalization to a unicode representation
// precent escapes for unreserved values are unescaped to their unicode value
// tests normalization to idna domains
// test ip word handling, ipv6 address handling, and trailing domain periods
// in general, this matches google chromes unescaping for things in the address bar.
// spaces are converted to '+' (perhaphs controversial)
// https://code.google.com/p/google-url/ probably is another good reference for this approach
func TestUrlnorm(t *testing.T) {
	testcases := map[string]string{
		"http://test.example/?a=%e3%82%82%26": "http://test.example/?a=%e3%82%82%26",
		//"http://test.example/?a=%e3%82%82%26": "http://test.example/?a=\xe3\x82\x82%26", //should return a unicode character
		"https://s.xn--q-bga.DE/":    "https://s.xn--q-bga.de/",       //should be in idna format
		"http://XBLA\u306eXbox.com": "https://xn--xblaxbox-jf4g.com", //test utf8 and unicode
		"http://президент.рф":       "http://kremlin.ru/",
		"http://ПРЕЗИДЕНТ.РФ":       "http://kremlin.ru/",
		"http://\u00e9.com":         "http://xn--9ca.com",
		"http://e\u0301.com":        "http://xn--9ca.com",
		"https://ja.wikipedia.org/wiki/%25E3%2582%25AD%25E3%2583%25A3%25E3%2582%25BF%25E3%2583%2594%25E3%2583%25A9%25E3%2583%25BC%25E3%2582%25B8%25E3%2583%25A3%25E3%2583%2591%25E3%2583%25B3": "https://ja.wikipedia.org/wiki/%25E3%2582%25AD%25E3%2583%25A3%25E3%2582%25BF%25E3%2583%2594%25E3%2583%25A9%25E3%2583%25BC%25E3%2582%25B8%25E3%2583%25A3%25E3%2583%2591%25E3%2583%25B3",
		//"https://ja.wikipedia.org/wiki/%25E3%2582%25AD%25E3%2583%25A3%25E3%2582%25BF%25E3%2583%2594%25E3%2583%25A9%25E3%2583%25BC%25E3%2582%25B8%25E3%2583%25A3%25E3%2583%2591%25E3%2583%25B3": "https://ja.wikipedia.org/wiki/\xe3\x82\xad\xe3\x83\xa3\xe3\x82\xbf\xe3\x83\x94\xe3\x83\xa9\xe3\x83\xbc\xe3\x82\xb8\xe3\x83\xa3\xe3\x83\x91\xe3\x83\xb3",

		"http://test.example/\xe3\x82\xad": "http://test.example/%E3%82%AD",
		//"http://test.example/\xe3\x82\xad":              "http://test.example/\xe3\x82\xad",
		"http://test.example/?p=%23val#test-%23-val%25": "http://test.example/?p=%23val#test-%23-val%25", //check that %23 (#) is not escaped where it shouldn't be

		"http://test.domain/I%C3%B1t%C3%ABrn%C3%A2ti%C3%B4n%EF%BF%BDliz%C3%A6ti%C3%B8n": "http://test.domain/I%C3%B1t%C3%ABrn%C3%A2ti%C3%B4n%EF%BF%BDliz%C3%A6ti%C3%B8n",
		//"http://test.domain/I%C3%B1t%C3%ABrn%C3%A2ti%C3%B4n%EF%BF%BDliz%C3%A6ti%C3%B8n": "http://test.domain/I\xc3\xb1t\xc3\xabrn\xc3\xa2ti\xc3\xb4n\xef\xbf\xbdliz\xc3\xa6ti\xc3\xb8n",
	}

	assertMap(t, testcases, FlagsSafe|FlagRemoveDotSegments)
}
