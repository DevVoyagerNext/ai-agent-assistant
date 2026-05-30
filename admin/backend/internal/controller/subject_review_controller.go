package controller

import (
	"admin/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

func (ctl *AdminController) PendingSubjects(c *gin.Context) {
	res, err := ctl.adminService.PendingSubjects()
	if err != nil {
		httpx.Fail(c, 500, err.Error())
		return
	}

	httpx.OK(c, res)
}

func (ctl *AdminController) ApproveSubject(c *gin.Context) {
	subjectID, ok := parseID(c)
	if !ok {
		httpx.Fail(c, 400, "教材 ID 不正确")
		return
	}

	req := bindReviewRequest(c)
	if err := ctl.adminService.ApproveSubject(currentAdminID(c), subjectID, req.Remark); err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, nil)
}

func (ctl *AdminController) RejectSubject(c *gin.Context) {
	subjectID, ok := parseID(c)
	if !ok {
		httpx.Fail(c, 400, "教材 ID 不正确")
		return
	}

	req := bindReviewRequest(c)
	if err := ctl.adminService.RejectSubject(currentAdminID(c), subjectID, req.Remark); err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, nil)
}
