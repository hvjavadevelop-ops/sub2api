package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelCatalogHandler struct {
	settingService *service.SettingService
}

func NewModelCatalogHandler(settingService *service.SettingService) *ModelCatalogHandler {
	return &ModelCatalogHandler{settingService: settingService}
}

func (h *ModelCatalogHandler) Get(c *gin.Context) {
	catalog, err := h.settingService.GetModelCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}

func (h *ModelCatalogHandler) Update(c *gin.Context) {
	var req service.ModelCatalogConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	catalog, err := h.settingService.UpdateModelCatalog(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}
