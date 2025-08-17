package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	http.HandleFunc("/debug/info", handleServiceInfo)

	fmt.Println("service run")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Fatal(err)
	}
}

func handleServiceInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("full") {
		fmt.Fprintln(w, buildInfo.String())
	} else {
		var vcsRevision, vcsTime string
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				vcsRevision = setting.Value
			case "vcs.time":
				vcsTime = setting.Value
			}
		}
		fmt.Fprintln(w, "Go Version:", buildInfo.GoVersion, "Build Version:", vcsRevision[:8], "Build Time:", vcsTime)
	}
}

var buildInfo *debug.BuildInfo = func() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return info
}()
