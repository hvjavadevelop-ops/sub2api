package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CreateTimedGrantRequest struct {
	GrantType       string  `json:"grant_type" binding:"required,oneof=balance concurrency"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	DurationSeconds int     `json:"duration_seconds" binding:"required,gt=0"`
	Notes           string  `json:"notes"`
}

func (h *UserHandler) CreateTimedGrant(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req CreateTimedGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var createdBy *int64
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok && subject.UserID > 0 {
		createdBy = &subject.UserID
	}
	grant, err := h.adminService.CreateTimedUserGrant(c.Request.Context(), userID, service.CreateTimedUserGrantInput{
		GrantType:       req.GrantType,
		Amount:          req.Amount,
		DurationSeconds: req.DurationSeconds,
		Notes:           req.Notes,
		CreatedBy:       createdBy,
	})
	if err != nil {
		if response.ErrorFrom(c, err) {
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create timed grant")
		return
	}
	response.Created(c, grant)
}

func (h *UserHandler) ListTimedGrants(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	grants, err := h.adminService.ListTimedUserGrants(c.Request.Context(), userID)
	if err != nil {
		if response.ErrorFrom(c, err) {
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to list timed grants")
		return
	}
	response.Success(c, grants)
}
