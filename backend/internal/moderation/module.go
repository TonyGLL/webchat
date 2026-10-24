// Package moderation implements roles management, kicks and the audit log.
package moderation

import (
	"backend/internal/moderation/application"
	"backend/internal/moderation/domain"
	"backend/internal/moderation/presentation"

	"github.com/gin-gonic/gin"
)

func RegisterModule(repo domain.Repository, realtime domain.Realtime, router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	controller := presentation.NewController(
		application.NewManageMembersUseCase(repo, realtime),
		application.NewListAuditUseCase(repo),
	)

	routes := router.Group("/moderation/rooms/:room_id")
	routes.Use(authMiddleware)
	{
		routes.PATCH("/members/:user_id/role", controller.ChangeRole)
		routes.DELETE("/members/:user_id", controller.Kick)
		routes.GET("/audit", controller.ListAudit)
	}
}
