package version

import (
	"runtime/debug"
)

var (
	version   string
	commit    string
	date      string
	goVersion string
	modified  bool
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("unable to retrieve build info")
	}
	version = bi.Main.Version
	goVersion = bi.GoVersion

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			commit = s.Value
		case "vcs.time":
			date = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
			version = version + "-dirty"
		default:
			continue
		}
	}
}

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func Date() string {
	return date
}

func GoVersion() string {
	return goVersion
}

func Modified() bool {
	return modified
}

func MarkModified(v *string) {
	if Modified() {
		*v = *v + "-dirty"
	}
}

func SetFromCmd(v string) {
	version = v
}

func Summary() string {
	return version + " (" + commit + ") (" + goVersion + ")"
}
