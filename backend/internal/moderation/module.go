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
		application.NewReportsUseCase(repo, realtime),
		application.NewBansUseCase(repo, realtime),
	)

	// Reports are filed against a message and resolved by id; the room is derived from them.
	global := router.Group("/moderation")
	global.Use(authMiddleware)
	{
		global.POST("/messages/:message_id/report", controller.Report)
		global.POST("/reports/:report_id/resolve", controller.ResolveReport)
	}

	routes := router.Group("/moderation/rooms/:room_id")
	routes.Use(authMiddleware)
	{
		routes.PATCH("/members/:user_id/role", controller.ChangeRole)
		routes.DELETE("/members/:user_id", controller.Kick)
		routes.GET("/audit", controller.ListAudit)
		routes.GET("/reports", controller.ListReports)
		routes.GET("/bans", controller.ListBans)
		routes.POST("/bans/:user_id", controller.Ban)
		routes.DELETE("/bans/:user_id", controller.Unban)
	}
}
