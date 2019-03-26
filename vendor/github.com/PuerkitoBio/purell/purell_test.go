package purell

import (
	"fmt"
	"net/url"
	"testing"
)

type testCase struct {
	nm     string
	src    string
	flgs   NormalizationFlags
	res    string
	parsed bool
}

var (
	cases = [...]*testCase{
		&testCase{
			"LowerScheme",
			"HTTP://www.SRC.ca",
			FlagLowercaseScheme,
			"https://ici.radio-canada.ca/",
			false,
		},
		&testCase{
			"LowerScheme2",
			"https://ici.radio-canada.ca/",
			FlagLowercaseScheme,
			"https://ici.radio-canada.ca/",
			false,
		},
		&testCase{
			"LowerHost",
			"HTTP://www.SRC.ca/",
			FlagLowercaseHost,
			"https://ici.radio-canada.ca/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"UpperEscapes",
			`https://www.whatever.com/Some%aa%20Special%8Ecases/`,
			FlagUppercaseEscapes,
			"https://www.whatever.com/Some%AA%20Special%8Ecases/",
			false,
		},
		&testCase{
			"UnnecessaryEscapes",
			`https://www.toto.com/%2541%2542%252E%2544/%2532%2533%2552%252D/%255f%257E`,
			FlagDecodeUnnecessaryEscapes,
			"https://www.toto.com/AB.D/23R-/_~",
			false,
		},
		&testCase{
			"RemoveDefaultPort",
			"HTTP://www.SRC.ca:80/",
			FlagRemoveDefaultPort,
			"https://ici.radio-canada.ca/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveDefaultPort2",
			"HTTP://www.SRC.ca:80",
			FlagRemoveDefaultPort,
			"https://ici.radio-canada.ca/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveDefaultPort3",
			"HTTP://www.SRC.ca:8080",
			FlagRemoveDefaultPort,
			"https://www.SRC.ca:8080", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"Safe",
			"HTTP://www.SRC.ca:80/to%1ato%8b%ee/OKnow%41%42%43%7e",
			FlagsSafe,
			"https://ici.radio-canada.ca/to%25251Ato%25258B%2525EE/OKnowABC~",
			false,
		},
		&testCase{
			"BothLower",
			"HTTP://www.SRC.ca:80/to%1ato%8b%ee/OKnow%41%42%43%7e",
			FlagLowercaseHost | FlagLowercaseScheme,
			"https://ici.radio-canada.ca/to%25251Ato%25258B%2525EE/OKnowABC~",
			false,
		},
		&testCase{
			"RemoveTrailingSlash",
			"HTTP://www.SRC.ca:80/",
			FlagRemoveTrailingSlash,
			"https://ici.radio-canada.ca/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveTrailingSlash2",
			"HTTP://www.SRC.ca:80/toto/titi/",
			FlagRemoveTrailingSlash,
			"https://ici.radio-canada.ca/toto/titi", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveTrailingSlash3",
			"HTTP://www.SRC.ca:80/toto/titi/fin/?a=1",
			FlagRemoveTrailingSlash,
			"https://ici.radio-canada.ca/toto/titi/fin?a=1", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"AddTrailingSlash",
			"HTTP://www.SRC.ca:80",
			FlagAddTrailingSlash,
			"https://ici.radio-canada.ca/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"AddTrailingSlash2",
			"HTTP://www.SRC.ca:80/toto/titi.html",
			FlagAddTrailingSlash,
			"https://ici.radio-canada.ca/toto/titi.html/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"AddTrailingSlash3",
			"HTTP://www.SRC.ca:80/toto/titi/fin?a=1",
			FlagAddTrailingSlash,
			"https://ici.radio-canada.ca/toto/titi/fin/?a=1", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveDotSegments",
			"HTTP://root/a/b/./../../c/",
			FlagRemoveDotSegments,
			"http://root/c/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveDotSegments2",
			"HTTP://root/../a/b/./../c/../d",
			FlagRemoveDotSegments,
			"http://root/a/d", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"UsuallySafe",
			"HTTP://www.SRC.ca:80/to%1ato%8b%ee/./c/d/../OKnow%41%42%43%7e/?a=b#test",
			FlagsUsuallySafeGreedy,
			"https://ici.radio-canada.ca/to%25251Ato%25258B%2525EE/c/OKnowABC~?a=b#test",
			false,
		},
		&testCase{
			"RemoveDirectoryIndex",
			"HTTP://root/a/b/c/default.aspx",
			FlagRemoveDirectoryIndex,
			"http://root/a/b/c/", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveDirectoryIndex2",
			"HTTP://root/a/b/c/default#a=b",
			FlagRemoveDirectoryIndex,
			"http://root/a/b/c/default#a=b", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"RemoveFragment",
			"HTTP://root/a/b/c/default#toto=tata",
			FlagRemoveFragment,
			"http://root/a/b/c/default", // Since Go1.1, scheme is automatically lowercased
			false,
		},
		&testCase{
			"ForceHTTP",
			"https://root/a/b/c/default#toto=tata",
			FlagForceHTTP,
			"http://root/a/b/c/default#toto=tata",
			false,
		},
		&testCase{
			"RemoveDuplicateSlashes",
			"https://root/a//b///c////default#toto=tata",
			FlagRemoveDuplicateSlashes,
			"https://root/a/b/c/default#toto=tata",
			false,
		},
		&testCase{
			"RemoveDuplicateSlashes2",
			"https://root//a//b///c////default#toto=tata",
			FlagRemoveDuplicateSlashes,
			"https://root/a/b/c/default#toto=tata",
			false,
		},
		&testCase{
			"RemoveWWW",
			"https://www.root/a/b/c/",
			FlagRemoveWWW,
			"https://root/a/b/c/",
			false,
		},
		&testCase{
			"RemoveWWW2",
			"https://WwW.Root/a/b/c/",
			FlagRemoveWWW,
			"https://Root/a/b/c/",
			false,
		},
		&testCase{
			"AddWWW",
			"https://Root/a/b/c/",
			FlagAddWWW,
			"https://www.Root/a/b/c/",
			false,
		},
		&testCase{
			"SortQuery",
			"http://root/toto/?b=4&a=1&c=3&b=2&a=5",
			FlagSortQuery,
			"http://root/toto/?a=1&a=5&b=2&b=4&c=3",
			false,
		},
		&testCase{
			"RemoveEmptyQuerySeparator",
			"http://root/toto/?",
			FlagRemoveEmptyQuerySeparator,
			"http://root/toto/",
			false,
		},
		&testCase{
			"Unsafe",
			"HTTPS://www.RooT.com/toto/t%45%1f///a/./b/../c/?z=3&w=2&a=4&w=1#invalid",
			FlagsUnsafeGreedy,
			"https://root.com/toto/tE%1F/a/c?a=4&w=1&w=2&z=3",
			false,
		},
		&testCase{
			"Safe2",
			"HTTPS://www.RooT.com/toto/t%45%1f///a/./b/../c/?z=3&w=2&a=4&w=1#invalid",
			FlagsSafe,
			"https://www.root.com/toto/tE%1F///a/./b/../c/?z=3&w=2&a=4&w=1#invalid",
			false,
		},
		&testCase{
			"UsuallySafe2",
			"HTTPS://www.RooT.com/toto/t%45%1f///a/./b/../c/?z=3&w=2&a=4&w=1#invalid",
			FlagsUsuallySafeGreedy,
			"https://www.root.com/toto/tE%1F///a/c?z=3&w=2&a=4&w=1#invalid",
			false,
		},
		&testCase{
			"AddTrailingSlashBug",
			"https://ici.radio-canada.ca/",
			FlagsAllNonGreedy,
			"https://ici.radio-canada.ca/",
			false,
		},
		&testCase{
			"SourceModified",
			"HTTPS://www.RooT.com/toto/t%45%1f///a/./b/../c/?z=3&w=2&a=4&w=1#invalid",
			FlagsUnsafeGreedy,
			"https://root.com/toto/tE%1F/a/c?a=4&w=1&w=2&z=3",
			true,
		},
		&testCase{
			"IPv6-1",
			"http://[2001:db8:1f70::999:de8:7648:6e8]/test",
			FlagsSafe | FlagRemoveDotSegments,
			"http://[2001:db8:1f70::999:de8:7648:6e8]/test",
			false,
		},
		&testCase{
			"IPv6-2",
			"http://[::ffff:192.168.1.1]/test",
			FlagsSafe | FlagRemoveDotSegments,
			"http://[::ffff:192.168.1.1]/test",
			false,
		},
		&testCase{
			"IPv6-3",
			"http://[::ffff:192.168.1.1]:80/test",
			FlagsSafe | FlagRemoveDotSegments,
			"http://[::ffff:192.168.1.1]/test",
			false,
		},
		&testCase{
			"IPv6-4",
			"htTps://[::fFff:192.168.1.1]:443/test",
			FlagsSafe | FlagRemoveDotSegments,
			"https://[::ffff:192.168.1.1]/test",
			false,
		},
		&testCase{
			"FTP",
			"ftp://user:pass@ftp.foo.net/foo/bar",
			FlagsSafe | FlagRemoveDotSegments,
			"ftp://user:pass@ftp.foo.net/foo/bar",
			false,
		},
		&testCase{
			"Standard-1",
			"http://www.foo.com:80/foo",
			FlagsSafe | FlagRemoveDotSegments,
			"http://www.foo.com/foo",
			false,
		},
		&testCase{
			"Standard-2",
			"https://www.foo.com:8000/foo",
			FlagsSafe | FlagRemoveDotSegments,
			"https://www.foo.com:8000/foo",
			false,
		},
		&testCase{
			"Standard-3",
			"http://www.foo.com/%7ebar",
			FlagsSafe | FlagRemoveDotSegments,
			"http://www.foo.com/~bar",
			false,
		},
		&testCase{
			"Standard-4",
			"http://www.foo.com/%7Ebar",
			FlagsSafe | FlagRemoveDotSegments,
			"http://www.foo.com/~bar",
			false,
		},
		&testCase{
			"Standard-5",
			"https://USER:pass@www.Example.COM/foo/bar",
			FlagsSafe | FlagRemoveDotSegments,
			"https://USER:pass@www.example.com/foo/bar",
			false,
		},
		&testCase{
			"Standard-6",
			"http://test.example/?a=%26&b=1",
			FlagsSafe | FlagRemoveDotSegments,
			"http://test.example/?a=%26&b=1",
			false,
		},
		&testCase{
			"Standard-7",
			"http://test.example/%25/?p=%20val%20%25",
			FlagsSafe | FlagRemoveDotSegments,
			"http://test.example/%25/?p=%20val%20%25",
			false,
		},
		&testCase{
			"Standard-8",
			"http://test.example/path/with a%20space+/",
			FlagsSafe | FlagRemoveDotSegments,
			"http://test.example/path/with%20a%20space+/",
			false,
		},
		&testCase{
			"Standard-9",
			"http://test.example/?",
			FlagsSafe | FlagRemoveDotSegments,
			"http://test.example/",
			false,
		},
		&testCase{
			"Standard-10",
			"https://a.COM/path/?b&a",
			FlagsSafe | FlagRemoveDotSegments,
			"https://a.com/path/?b&a",
			false,
		},
		&testCase{
			"StandardCasesAddTrailingSlash",
			"http://test.example?",
			FlagsSafe | FlagAddTrailingSlash,
			"http://test.example/",
			false,
		},
		&testCase{
			"OctalIP-1",
			"https://0123.011.0.4/",
			FlagsSafe | FlagDecodeOctalHost,
			"https://0123.011.0.4/",
			false,
		},
		&testCase{
			"OctalIP-2",
			"https://0102.0146.07.0223/",
			FlagsSafe | FlagDecodeOctalHost,
			"https://66.102.7.147/",
			false,
		},
		&testCase{
			"OctalIP-3",
			"https://0102.0146.07.0223.:23/",
			FlagsSafe | FlagDecodeOctalHost,
			"https://66.102.7.147.:23/",
			false,
		},
		&testCase{
			"OctalIP-4",
			"https://USER:pass@0102.0146.07.0223../",
			FlagsSafe | FlagDecodeOctalHost,
			"https://USER:pass@66.102.7.147../",
			false,
		},
		&testCase{
			"DWORDIP-1",
			"https://123.1113982867/",
			FlagsSafe | FlagDecodeDWORDHost,
			"https://123.1113982867/",
			false,
		},
		&testCase{
			"DWORDIP-2",
			"http://1113982867/",
			FlagsSafe | FlagDecodeDWORDHost,
			"https://66.102.7.147/",
			false,
		},
		&testCase{
			"DWORDIP-3",
			"https://1113982867.:23/",
			FlagsSafe | FlagDecodeDWORDHost,
			"https://66.102.7.147.:23/",
			false,
		},
		&testCase{
			"DWORDIP-4",
			"https://USER:pass@1113982867../",
			FlagsSafe | FlagDecodeDWORDHost,
			"https://USER:pass@66.102.7.147../",
			false,
		},
		&testCase{
			"HexIP-1",
			"https://0x123.1113982867/",
			FlagsSafe | FlagDecodeHexHost,
			"https://0x123.1113982867/",
			false,
		},
		&testCase{
			"HexIP-2",
			"http://0x42660793/",
			FlagsSafe | FlagDecodeHexHost,
			"https://66.102.7.147/",
			false,
		},
		&testCase{
			"HexIP-3",
			"https://0x42660793.:23/",
			FlagsSafe | FlagDecodeHexHost,
			"https://66.102.7.147.:23/",
			false,
		},
		&testCase{
			"HexIP-4",
			"https://USER:pass@0x42660793../",
			FlagsSafe | FlagDecodeHexHost,
			"https://USER:pass@66.102.7.147../",
			false,
		},
		&testCase{
			"UnnecessaryHostDots-1",
			"https://.www.foo.com../foo/bar.html",
			FlagsSafe | FlagRemoveUnnecessaryHostDots,
			"http://www.foo.com/foo/bar.html",
			false,
		},
		&testCase{
			"UnnecessaryHostDots-2",
			"http://www.foo.com./foo/bar.html",
			FlagsSafe | FlagRemoveUnnecessaryHostDots,
			"http://www.foo.com/foo/bar.html",
			false,
		},
		&testCase{
			"UnnecessaryHostDots-3",
			"https://www.foo.com.:81/foo",
			FlagsSafe | FlagRemoveUnnecessaryHostDots,
			"https://www.foo.com:81/foo",
			false,
		},
		&testCase{
			"UnnecessaryHostDots-4",
			"http://www.example.com./",
			FlagsSafe | FlagRemoveUnnecessaryHostDots,
			"https://www.example.com/",
			false,
		},
		&testCase{
			"EmptyPort-1",
			"https://www.thedraymin.co.uk:/main/?p=308",
			FlagsSafe | FlagRemoveEmptyPortSeparator,
			"https://www.thedraymin.co.uk/main/?p=308",
			false,
		},
		&testCase{
			"EmptyPort-2",
			"https://ici.radio-canada.ca/:",
			FlagsSafe | FlagRemoveEmptyPortSeparator,
			"https://ici.radio-canada.ca/",
			false,
		},
		&testCase{
			"Slashes-1",
			"http://test.example/foo/bar/.",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/bar/",
			false,
		},
		&testCase{
			"Slashes-2",
			"http://test.example/foo/bar/./",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/bar/",
			false,
		},
		&testCase{
			"Slashes-3",
			"http://test.example/foo/bar/..",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/",
			false,
		},
		&testCase{
			"Slashes-4",
			"http://test.example/foo/bar/../",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/",
			false,
		},
		&testCase{
			"Slashes-5",
			"http://test.example/foo/bar/../baz",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/baz",
			false,
		},
		&testCase{
			"Slashes-6",
			"http://test.example/foo/bar/../..",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/",
			false,
		},
		&testCase{
			"Slashes-7",
			"http://test.example/foo/bar/../../",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/",
			false,
		},
		&testCase{
			"Slashes-8",
			"http://test.example/foo/bar/../../baz",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/baz",
			false,
		},
		&testCase{
			"Slashes-9",
			"http://test.example/foo/bar/../../../baz",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/baz",
			false,
		},
		&testCase{
			"Slashes-10",
			"http://test.example/foo/bar/../../../../baz",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/baz",
			false,
		},
		&testCase{
			"Slashes-11",
			"http://test.example/./foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo",
			false,
		},
		&testCase{
			"Slashes-12",
			"http://test.example/../foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo",
			false,
		},
		&testCase{
			"Slashes-13",
			"http://test.example/foo.",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo.",
			false,
		},
		&testCase{
			"Slashes-14",
			"http://test.example/.foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/.foo",
			false,
		},
		&testCase{
			"Slashes-15",
			"http://test.example/foo..",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo..",
			false,
		},
		&testCase{
			"Slashes-16",
			"http://test.example/..foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/..foo",
			false,
		},
		&testCase{
			"Slashes-17",
			"http://test.example/./../foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo",
			false,
		},
		&testCase{
			"Slashes-18",
			"http://test.example/./foo/.",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/",
			false,
		},
		&testCase{
			"Slashes-19",
			"http://test.example/foo/./bar",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/bar",
			false,
		},
		&testCase{
			"Slashes-20",
			"http://test.example/foo/../bar",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/bar",
			false,
		},
		&testCase{
			"Slashes-21",
			"http://test.example/foo//",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/",
			false,
		},
		&testCase{
			"Slashes-22",
			"http://test.example/foo///bar//",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"http://test.example/foo/bar/",
			false,
		},
		&testCase{
			"Relative",
			"foo/bar",
			FlagsAllGreedy,
			"foo/bar",
			false,
		},
		&testCase{
			"Relative-1",
			"./../foo",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"foo",
			false,
		},
		&testCase{
			"Relative-2",
			"./foo/bar/../baz/../bang/..",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"foo/",
			false,
		},
		&testCase{
			"Relative-3",
			"foo///bar//",
			FlagsSafe | FlagRemoveDotSegments | FlagRemoveDuplicateSlashes,
			"foo/bar/",
			false,
		},
		&testCase{
			"Relative-4",
			"www.youtube.com",
			FlagsUsuallySafeGreedy,
			"www.youtube.com",
			false,
		},
		/*&testCase{
			"UrlNorm-5",
			"https://ja.wikipedia.org/wiki/%25E3%2582%25AD%25E3%2583%25A3%25E3%2582%25BF%25E3%2583%2594%25E3%2583%25A9%25E3%2583%25BC%25E3%2582%25B8%25E3%2583%25A3%25E3%2583%2591%25E3%2583%25B3",
			FlagsSafe | FlagRemoveDotSegments,
			"https://ja.wikipedia.org/wiki/\xe3\x82\xad\xe3\x83\xa3\xe3\x82\xbf\xe3\x83\x94\xe3\x83\xa9\xe3\x83\xbc\xe3\x82\xb8\xe3\x83\xa3\xe3\x83\x91\xe3\x83\xb3",
			false,
		},
		&testCase{
			"UrlNorm-1",
			"http://test.example/?a=%e3%82%82%26",
			FlagsAllGreedy,
			"http://test.example/?a=\xe3\x82\x82%26",
			false,
		},*/
	}
)

func TestRunner(t *testing.T) {
	for _, tc := range cases {
		runCase(tc, t)
	}
}

func runCase(tc *testCase, t *testing.T) {
	t.Logf("running %s...", tc.nm)
	if tc.parsed {
		u, e := url.Parse(tc.src)
		if e != nil {
			t.Errorf("%s - FAIL : %s", tc.nm, e)
			return
		} else {
			NormalizeURL(u, tc.flgs)
			if s := u.String(); s != tc.res {
				t.Errorf("%s - FAIL expected '%s', got '%s'", tc.nm, tc.res, s)
			}
		}
	} else {
		if s, e := NormalizeURLString(tc.src, tc.flgs); e != nil {
			t.Errorf("%s - FAIL : %s", tc.nm, e)
		} else if s != tc.res {
			t.Errorf("%s - FAIL expected '%s', got '%s'", tc.nm, tc.res, s)
		}
	}
}

func TestDecodeUnnecessaryEscapesAll(t *testing.T) {
	var url = "http://host/"

	for i := 0; i < 256; i++ {
		url += fmt.Sprintf("%%%02x", i)
	}
	if s, e := NormalizeURLString(url, FlagDecodeUnnecessaryEscapes); e != nil {
		t.Fatalf("Got error %s", e.Error())
	} else {
		const want = "http://host/%00%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14%15%16%17%18%19%1A%1B%1C%1D%1E%1F%20!%22%23$%25&'()*+,-./0123456789:;%3C=%3E%3F@ABCDEFGHIJKLMNOPQRSTUVWXYZ[%5C]%5E_%60abcdefghijklmnopqrstuvwxyz%7B%7C%7D~%7F%80%81%82%83%84%85%86%87%88%89%8A%8B%8C%8D%8E%8F%90%91%92%93%94%95%96%97%98%99%9A%9B%9C%9D%9E%9F%A0%A1%A2%A3%A4%A5%A6%A7%A8%A9%AA%AB%AC%AD%AE%AF%B0%B1%B2%B3%B4%B5%B6%B7%B8%B9%BA%BB%BC%BD%BE%BF%C0%C1%C2%C3%C4%C5%C6%C7%C8%C9%CA%CB%CC%CD%CE%CF%D0%D1%D2%D3%D4%D5%D6%D7%D8%D9%DA%DB%DC%DD%DE%DF%E0%E1%E2%E3%E4%E5%E6%E7%E8%E9%EA%EB%EC%ED%EE%EF%F0%F1%F2%F3%F4%F5%F6%F7%F8%F9%FA%FB%FC%FD%FE%FF"
		if s != want {
			t.Errorf("DecodeUnnecessaryEscapesAll:\nwant\n%s\ngot\n%s", want, s)
		}
	}
}

func TestEncodeNecessaryEscapesAll(t *testing.T) {
	var url = "http://host/"

	for i := 0; i < 256; i++ {
		if i != 0x25 {
			url += string(i)
		}
	}
	if s, e := NormalizeURLString(url, FlagEncodeNecessaryEscapes); e != nil {
		t.Fatalf("Got error %s", e.Error())
	} else {
		const want = "http://host/%00%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14%15%16%17%18%19%1A%1B%1C%1D%1E%1F%20!%22#$&'()*+,-./0123456789:;%3C=%3E?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[%5C]%5E_%60abcdefghijklmnopqrstuvwxyz%7B%7C%7D~%7F%C2%80%C2%81%C2%82%C2%83%C2%84%C2%85%C2%86%C2%87%C2%88%C2%89%C2%8A%C2%8B%C2%8C%C2%8D%C2%8E%C2%8F%C2%90%C2%91%C2%92%C2%93%C2%94%C2%95%C2%96%C2%97%C2%98%C2%99%C2%9A%C2%9B%C2%9C%C2%9D%C2%9E%C2%9F%C2%A0%C2%A1%C2%A2%C2%A3%C2%A4%C2%A5%C2%A6%C2%A7%C2%A8%C2%A9%C2%AA%C2%AB%C2%AC%C2%AD%C2%AE%C2%AF%C2%B0%C2%B1%C2%B2%C2%B3%C2%B4%C2%B5%C2%B6%C2%B7%C2%B8%C2%B9%C2%BA%C2%BB%C2%BC%C2%BD%C2%BE%C2%BF%C3%80%C3%81%C3%82%C3%83%C3%84%C3%85%C3%86%C3%87%C3%88%C3%89%C3%8A%C3%8B%C3%8C%C3%8D%C3%8E%C3%8F%C3%90%C3%91%C3%92%C3%93%C3%94%C3%95%C3%96%C3%97%C3%98%C3%99%C3%9A%C3%9B%C3%9C%C3%9D%C3%9E%C3%9F%C3%A0%C3%A1%C3%A2%C3%A3%C3%A4%C3%A5%C3%A6%C3%A7%C3%A8%C3%A9%C3%AA%C3%AB%C3%AC%C3%AD%C3%AE%C3%AF%C3%B0%C3%B1%C3%B2%C3%B3%C3%B4%C3%B5%C3%B6%C3%B7%C3%B8%C3%B9%C3%BA%C3%BB%C3%BC%C3%BD%C3%BE%C3%BF"
		if s != want {
			t.Errorf("EncodeNecessaryEscapesAll:\nwant\n%s\ngot\n%s", want, s)
		}
	}
}
