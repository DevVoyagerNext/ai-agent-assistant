package course

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	SubjectsApi
}

var subjectsService = service.ServiceGroupApp.CourseServiceGroup.SubjectsService
