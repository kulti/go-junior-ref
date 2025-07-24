package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func main() {
	http.HandleFunc("/debug/info", handleServiceInfo)

	fmt.Println("service run")
	http.ListenAndServe(":8090", nil)
}

func handleServiceInfo(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, buildInfo.String())
}

var buildInfo *debug.BuildInfo = func() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return info
}()
