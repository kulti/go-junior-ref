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

func handleServiceInfo(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, buildInfo.String())
}

var buildInfo *debug.BuildInfo = func() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return info
}()
