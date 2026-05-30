package controller

import (
	"admin/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

func (ctl *AdminController) PendingNodes(c *gin.Context) {
	res, err := ctl.adminService.PendingNodes()
	if err != nil {
		httpx.Fail(c, 500, err.Error())
		return
	}

	httpx.OK(c, res)
}

func (ctl *AdminController) ApproveNode(c *gin.Context) {
	nodeID, ok := parseID(c)
	if !ok {
		httpx.Fail(c, 400, "节点 ID 不正确")
		return
	}

	req := bindReviewRequest(c)
	if err := ctl.adminService.ApproveNode(currentAdminID(c), nodeID, req.Remark); err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, nil)
}

func (ctl *AdminController) RejectNode(c *gin.Context) {
	nodeID, ok := parseID(c)
	if !ok {
		httpx.Fail(c, 400, "节点 ID 不正确")
		return
	}

	req := bindReviewRequest(c)
	if err := ctl.adminService.RejectNode(currentAdminID(c), nodeID, req.Remark); err != nil {
		httpx.Fail(c, 400, err.Error())
		return
	}

	httpx.OK(c, nil)
}
