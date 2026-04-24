package course

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/course"
    courseReq "github.com/flipped-aurora/gin-vue-admin/server/model/course/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SubjectsApi struct {}



// CreateSubjects 创建教材审批
// @Tags Subjects
// @Summary 创建教材审批
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body course.Subjects true "创建教材审批"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /subjects/createSubjects [post]
func (subjectsApi *SubjectsApi) CreateSubjects(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var subjects course.Subjects
	err := c.ShouldBindJSON(&subjects)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = subjectsService.CreateSubjects(ctx,&subjects)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSubjects 删除教材审批
// @Tags Subjects
// @Summary 删除教材审批
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body course.Subjects true "删除教材审批"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /subjects/deleteSubjects [delete]
func (subjectsApi *SubjectsApi) DeleteSubjects(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := subjectsService.DeleteSubjects(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSubjectsByIds 批量删除教材审批
// @Tags Subjects
// @Summary 批量删除教材审批
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /subjects/deleteSubjectsByIds [delete]
func (subjectsApi *SubjectsApi) DeleteSubjectsByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := subjectsService.DeleteSubjectsByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSubjects 更新教材审批
// @Tags Subjects
// @Summary 更新教材审批
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body course.Subjects true "更新教材审批"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /subjects/updateSubjects [put]
func (subjectsApi *SubjectsApi) UpdateSubjects(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var subjects course.Subjects
	err := c.ShouldBindJSON(&subjects)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = subjectsService.UpdateSubjects(ctx,subjects)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSubjects 用id查询教材审批
// @Tags Subjects
// @Summary 用id查询教材审批
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询教材审批"
// @Success 200 {object} response.Response{data=course.Subjects,msg=string} "查询成功"
// @Router /subjects/findSubjects [get]
func (subjectsApi *SubjectsApi) FindSubjects(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resubjects, err := subjectsService.GetSubjects(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resubjects, c)
}
// GetSubjectsList 分页获取教材审批列表
// @Tags Subjects
// @Summary 分页获取教材审批列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query courseReq.SubjectsSearch true "分页获取教材审批列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /subjects/getSubjectsList [get]
func (subjectsApi *SubjectsApi) GetSubjectsList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo courseReq.SubjectsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := subjectsService.GetSubjectsInfoList(ctx,pageInfo)
	if err != nil {
	    global.GVA_LOG.Error("获取失败!", zap.Error(err))
        response.FailWithMessage("获取失败:" + err.Error(), c)
        return
    }
    response.OkWithDetailed(response.PageResult{
        List:     list,
        Total:    total,
        Page:     pageInfo.Page,
        PageSize: pageInfo.PageSize,
    }, "获取成功", c)
}

// GetSubjectsPublic 不需要鉴权的教材审批接口
// @Tags Subjects
// @Summary 不需要鉴权的教材审批接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /subjects/getSubjectsPublic [get]
func (subjectsApi *SubjectsApi) GetSubjectsPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    subjectsService.GetSubjectsPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的教材审批接口信息",
    }, "获取成功", c)
}
