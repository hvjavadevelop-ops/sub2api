package handler

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

func (h *ModelCatalogHandler) GetPublic(c *gin.Context) {
	catalog, err := h.settingService.GetModelCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}
