package main

import (
	"sort"
)

//nolint:gochecknoglobals // read-only variable
var servicePorts = map[string]int{
	"api-server": 8090,
}

func allServices() []string {
	services := make([]string, 0, len(servicePorts))
	for name := range servicePorts {
		services = append(services, name)
	}
	sort.Strings(services)
	return services
}
