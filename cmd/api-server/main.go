package main

import (
	"github.com/kulti/task_list_course/internal/app/apiserver"
	"github.com/kulti/task_list_course/internal/service"
)

func main() {
	service.Main(apiserver.New)
}
