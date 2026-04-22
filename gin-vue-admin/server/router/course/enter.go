package course

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	SubjectsRouter
}

var subjectsApi = api.ApiGroupApp.CourseApiGroup.SubjectsApi
