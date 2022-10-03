package selfupdater

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
)

const (
	debRE string = `(?P<pre>^deb(-src)?.* )(?P<url>https://packagecloud\.io/tsuru/\w+/)(?P<os>\w+)(?P<sep>/? )(?P<dist>[0-9A-Za-z.]+)(?P<end> main.*$)`
	rpmRE string = `(?P<pre>^baseurl=)(?P<url>https://packagecloud\.io/tsuru/\w+/)(?P<os>\w+)(?P<sep>/)(?P<dist>[0-9A-Za-z.]+)(?P<end>/.*$)`
)

func reFindSubmatchMap(r *regexp.Regexp, data string) map[string]string {
	match := r.FindStringSubmatch(data)
	matchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && i < len(match) {
			matchMap[name] = match[i]
		}
	}
	return matchMap
}

func findConfRepoPath() (string, string) {
	if _, err := os.Stat("/etc/apt/sources.list.d/tsuru_stable.list"); err == nil {
		return "deb", "/etc/apt/sources.list.d/tsuru_stable.list"
	}
	if _, err := os.Stat("/etc/zypp/repos.d/tsuru_stable.repo"); err == nil {
		return "rpm", "/etc/zypp/repos.d/tsuru_stable.repo"
	}
	if _, err := os.Stat("/etc/yum.repos.d/tsuru_stable.repo"); err == nil {
		return "rpm", "/etc/yum.repos.d/tsuru_stable.repo"
	}
	return "", ""
}

// replaceConfLine checks line with regex r.
func replaceConfLine(r *regexp.Regexp, line string) (wasReplaced bool, replacedLine string) {
	m := reFindSubmatchMap(r, line)
	if len(m) > 0 {
		for _, k := range []string{"pre", "url", "os", "sep", "dist", "end"} {
			if v, _ := m[k]; v == "" {
				return false, line
			}
		}
		if m["os"] != "any" || m["dist"] != "any" {
			// was:         pre   +    url   +  os   +    sep   + dist  +    end
			return true, m["pre"] + m["url"] + "any" + m["sep"] + "any" + m["end"]
		}
	}
	return false, line
}

func replaceConf(r *regexp.Regexp, reader io.Reader) (hasDiff bool, replacedContent *bytes.Buffer, err error) {
	scanner := bufio.NewScanner(reader)
	writer := &bytes.Buffer{}
	for scanner.Scan() {
		wasReplaced, line := replaceConfLine(r, scanner.Text())
		if wasReplaced {
			hasDiff = true
		}
		writer.WriteString(line + "\n")
	}
	if err = scanner.Err(); err != nil {
		return hasDiff, writer, fmt.Errorf("Got error on scanning repoConfPath lines: %w", err)
	}
	return hasDiff, writer, err
}

func checkUpToDateConfRepo(repoType, repoConfPath string) error {
	var r *regexp.Regexp

	switch repoType {
	case "deb":
		r = regexp.MustCompile(debRE)
	case "rpm":
		r = regexp.MustCompile(rpmRE)
	default:
		return nil
	}

	file, err := os.Open(repoConfPath)
	if err != nil {
		return fmt.Errorf("Could not open repoConfPath: %w", err)
	}
	defer file.Close()

	hasDiff, newContent, err := replaceConf(r, file)

	// XXX: TODO

	fmt.Println(newContent, hasDiff)
	return nil
}

func CheckPackageCloudRepo() {
	repoType, repoConfPath := findConfRepoPath()
	checkUpToDateConfRepo(repoType, repoConfPath)
}
