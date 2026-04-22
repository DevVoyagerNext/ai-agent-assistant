package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service/course"
	"github.com/flipped-aurora/gin-vue-admin/server/service/example"
	"github.com/flipped-aurora/gin-vue-admin/server/service/study"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
	StudyServiceGroup   study.ServiceGroup
	CourseServiceGroup  course.ServiceGroup
}
