import service from '@/utils/request'
// @Tags Subjects
// @Summary 创建subjects表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Subjects true "创建subjects表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /subjects/createSubjects [post]
export const createSubjects = (data) => {
  return service({
    url: '/subjects/createSubjects',
    method: 'post',
    data
  })
}

// @Tags Subjects
// @Summary 删除subjects表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Subjects true "删除subjects表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /subjects/deleteSubjects [delete]
export const deleteSubjects = (params) => {
  return service({
    url: '/subjects/deleteSubjects',
    method: 'delete',
    params
  })
}

// @Tags Subjects
// @Summary 批量删除subjects表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除subjects表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /subjects/deleteSubjects [delete]
export const deleteSubjectsByIds = (params) => {
  return service({
    url: '/subjects/deleteSubjectsByIds',
    method: 'delete',
    params
  })
}

// @Tags Subjects
// @Summary 更新subjects表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.Subjects true "更新subjects表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /subjects/updateSubjects [put]
export const updateSubjects = (data) => {
  return service({
    url: '/subjects/updateSubjects',
    method: 'put',
    data
  })
}

// @Tags Subjects
// @Summary 用id查询subjects表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.Subjects true "用id查询subjects表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /subjects/findSubjects [get]
export const findSubjects = (params) => {
  return service({
    url: '/subjects/findSubjects',
    method: 'get',
    params
  })
}

// @Tags Subjects
// @Summary 分页获取subjects表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取subjects表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /subjects/getSubjectsList [get]
export const getSubjectsList = (params) => {
  return service({
    url: '/subjects/getSubjectsList',
    method: 'get',
    params
  })
}

// @Tags Subjects
// @Summary 不需要鉴权的subjects表接口
// @Accept application/json
// @Produce application/json
// @Param data query courseReq.SubjectsSearch true "分页获取subjects表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /subjects/getSubjectsPublic [get]
export const getSubjectsPublic = () => {
  return service({
    url: '/subjects/getSubjectsPublic',
    method: 'get',
  })
}
