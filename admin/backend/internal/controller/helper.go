package controller

import (
	"admin/backend/internal/dto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func bindReviewRequest(c *gin.Context) dto.ReviewRequest {
	var req dto.ReviewRequest
	_ = c.ShouldBindJSON(&req)
	req.Remark = strings.TrimSpace(req.Remark)
	return req
}

func currentAdminID(c *gin.Context) uint {
	value, ok := c.Get("adminId")
	if !ok {
		return 0
	}
	adminID, _ := value.(uint)
	return adminID
}

func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	return id, err == nil && id > 0
}
