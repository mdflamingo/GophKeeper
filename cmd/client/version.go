package main

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

type Info struct {
	Version   string
	BuildDate string
	GitCommit string
	GoVersion string
	Platform  string
}

func GetInfo() Info {
	return Info{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (i Info) String() string {
	return fmt.Sprintf(
		"Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s\nPlatform: %s",
		i.Version, i.BuildDate, i.GitCommit, i.GoVersion, i.Platform,
	)
}

func (i Info) Short() string {
	return fmt.Sprintf("%s (%s)", i.Version, i.BuildDate)
}
