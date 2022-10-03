package selfupdater

import (
	"regexp"
	"strings"

	"gopkg.in/check.v1"
)

const (
	debSourceList = `# this file was generated by packagecloud.io for
# the repository at https://packagecloud.io/tsuru/stable

deb [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] https://packagecloud.io/tsuru/stable/ubuntu/ bionic main
deb-src [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] https://packagecloud.io/tsuru/stable/ubuntu/ bionic main
`
	rpmSourceList = `[tsuru_stable]
name=tsuru_stable
baseurl=https://packagecloud.io/tsuru/stable/opensuse/13.1/$basearch
repo_gpgcheck=1
pkg_gpgcheck=0
enabled=1
gpgkey=https://packagecloud.io/tsuru/stable/gpgkey
autorefresh=1
type=rpm-md

[tsuru_stable-source]
name=tsuru_stable-source
baseurl=https://packagecloud.io/tsuru/stable/opensuse/13.1/SRPMS
repo_gpgcheck=1
pkg_gpgcheck=0
enabled=1
gpgkey=https://packagecloud.io/tsuru/stable/gpgkey
autorefresh=1
type=rpm-md
`
)

func (s *S) TestReSubMatchMapDeb(c *check.C) {
	r := regexp.MustCompile(debRE)

	data := "deb [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] https://packagecloud.io/tsuru/stable/debian/ buster main"
	match := reFindSubmatchMap(r, data)
	c.Assert(string(match["pre"]), check.Equals, "deb [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] ")
	c.Assert(string(match["url"]), check.Equals, "https://packagecloud.io/tsuru/stable/")
	c.Assert(string(match["os"]), check.Equals, "debian")
	c.Assert(string(match["sep"]), check.Equals, "/ ")
	c.Assert(string(match["dist"]), check.Equals, "buster")
	c.Assert(string(match["end"]), check.Equals, " main")

	data = "deb-src https://packagecloud.io/tsuru/stable/debian/ buster main"
	match = reFindSubmatchMap(r, data)
	c.Assert(string(match["pre"]), check.Equals, "deb-src ")
	c.Assert(string(match["url"]), check.Equals, "https://packagecloud.io/tsuru/stable/")
	c.Assert(string(match["os"]), check.Equals, "debian")
	c.Assert(string(match["sep"]), check.Equals, "/ ")
	c.Assert(string(match["dist"]), check.Equals, "buster")
	c.Assert(string(match["end"]), check.Equals, " main")
}

func (s *S) TestReSubMatchMapRpm(c *check.C) {
	r := regexp.MustCompile(rpmRE)

	data := "baseurl=https://packagecloud.io/tsuru/stable/opensuse/13.1/$basearch"
	match := reFindSubmatchMap(r, data)
	c.Assert(string(match["pre"]), check.Equals, "baseurl=")
	c.Assert(string(match["url"]), check.Equals, "https://packagecloud.io/tsuru/stable/")
	c.Assert(string(match["os"]), check.Equals, "opensuse")
	c.Assert(string(match["sep"]), check.Equals, "/")
	c.Assert(string(match["dist"]), check.Equals, "13.1")
	c.Assert(string(match["end"]), check.Equals, "/$basearch")
}

func (s *S) TestReplaceConfLine(c *check.C) {
	for _, testCase := range []struct {
		r           *regexp.Regexp
		inputLine   string
		line        string
		wasReplaced bool
	}{
		// Testing DEB
		{
			regexp.MustCompile(debRE),
			"deb https://packagecloud.io/tsuru/stable/debian/ buster main",
			"deb https://packagecloud.io/tsuru/stable/any/ any main",
			true,
		},
		{
			regexp.MustCompile(debRE),
			"deb-src https://packagecloud.io/tsuru/stable/any/ any main",
			"deb-src https://packagecloud.io/tsuru/stable/any/ any main",
			false,
		},
		{
			regexp.MustCompile(debRE),
			"deb [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] https://packagecloud.io/tsuru/stable/someOs/ someDist main",
			"deb [signed-by=/etc/apt/keyrings/tsuru_stable-archive-keyring.gpg] https://packagecloud.io/tsuru/stable/any/ any main",
			true,
		},
		{
			regexp.MustCompile(debRE),
			"no match at all",
			"no match at all",
			false,
		},
		{
			regexp.MustCompile(debRE),
			"deb https://packagecloud.io/tsuru/someChannel/debian/ buster main",
			"deb https://packagecloud.io/tsuru/someChannel/any/ any main",
			true,
		},
		{
			regexp.MustCompile(debRE),
			"deb https://almostvalid.but.not/tsuru/stable/debian/ buster main",
			"deb https://almostvalid.but.not/tsuru/stable/debian/ buster main",
			false,
		},
		{
			// should be ignored
			regexp.MustCompile(debRE),
			"#deb https://packagecloud.io/tsuru/stable/debian/ buster main",
			"#deb https://packagecloud.io/tsuru/stable/debian/ buster main",
			false,
		},
		// Testing RPM
		{
			regexp.MustCompile(rpmRE),
			"baseurl=https://packagecloud.io/tsuru/stable/opensuse/13.1/$basearch",
			"baseurl=https://packagecloud.io/tsuru/stable/any/any/$basearch",
			true,
		},
		{
			regexp.MustCompile(rpmRE),
			"baseurl=https://packagecloud.io/tsuru/stable/opensuse/13.1/SRPMS",
			"baseurl=https://packagecloud.io/tsuru/stable/any/any/SRPMS",
			true,
		},
	} {
		wasReplaced, line := replaceConfLine(testCase.r, testCase.inputLine)
		c.Assert(wasReplaced, check.Equals, testCase.wasReplaced)
		c.Assert(line, check.Equals, testCase.line)
	}
}

func (s *S) TestReplaceConf(c *check.C) {
	hasDiff, newContent, err := replaceConf(regexp.MustCompile(debRE), strings.NewReader(debSourceList))
	expected := strings.ReplaceAll(debSourceList, "ubuntu", "any")
	expected = strings.ReplaceAll(expected, "bionic", "any")
	c.Assert(hasDiff, check.Equals, true)
	c.Assert(newContent.String(), check.Equals, expected)
	c.Assert(err, check.IsNil)

	hasDiff, newContent, err = replaceConf(regexp.MustCompile(rpmRE), strings.NewReader(rpmSourceList))
	expected = strings.ReplaceAll(rpmSourceList, "opensuse", "any")
	expected = strings.ReplaceAll(expected, "13.1", "any")
	c.Assert(hasDiff, check.Equals, true)
	c.Assert(newContent.String(), check.Equals, expected)
	c.Assert(err, check.IsNil)
}
