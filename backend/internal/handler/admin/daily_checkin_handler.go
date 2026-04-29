package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DailyCheckinHandler struct {
	service *service.DailyCheckinService
}

func NewDailyCheckinHandler(service *service.DailyCheckinService) *DailyCheckinHandler {
	return &DailyCheckinHandler{service: service}
}

func (h *DailyCheckinHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = parsed
	}
	records, total, err := h.service.ListAdmin(c.Request.Context(), service.DailyCheckinListParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		UserID:    userID,
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, records, int64(total), page, pageSize)
}
