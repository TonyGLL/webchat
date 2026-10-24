package presentation

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend/internal/moderation/application"
	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	members *application.ManageMembersUseCase
	audit   *application.ListAuditUseCase
	reports *application.ReportsUseCase
	bans    *application.BansUseCase
}

func NewController(members *application.ManageMembersUseCase, audit *application.ListAuditUseCase, reports *application.ReportsUseCase, bans *application.BansUseCase) *Controller {
	return &Controller{members: members, audit: audit, reports: reports, bans: bans}
}

// ids reads the authenticated user and the :room_id / :user_id path parameters.
func ids(c *gin.Context) (actorID int, roomID uuid.UUID, targetID int, ok bool) {
	actorID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	roomID, err = uuid.Parse(c.Param("room_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	if raw := c.Param("user_id"); raw != "" {
		targetID, err = strconv.Atoi(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid user ID")
			return
		}
	}
	return actorID, roomID, targetID, true
}

// respondError maps domain errors to HTTP statuses without leaking internals.
func respondError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, shared_domain.ErrForbidden), errors.Is(err, domain.ErrCannotTargetOwner):
		response.Error(c, http.StatusForbidden, "You are not allowed to do this")
	case errors.Is(err, domain.ErrNotMember), errors.Is(err, shared_domain.ErrNotFound),
		errors.Is(err, domain.ErrReportNotFound), errors.Is(err, domain.ErrMessageNotFound):
		response.Error(c, http.StatusNotFound, "Not found")
	case errors.Is(err, shared_domain.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "Invalid input")
	case errors.Is(err, shared_domain.ErrConflict), errors.Is(err, domain.ErrAlreadyReported):
		response.Error(c, http.StatusConflict, "Already done")
	default:
		log.Printf("%s: %v", fallback, err)
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}

type roleBody struct {
	Role string `json:"role" binding:"required"`
}

// ChangeRole handles PATCH /moderation/rooms/:room_id/members/:user_id/role
func (ctrl *Controller) ChangeRole(c *gin.Context) {
	actorID, roomID, targetID, ok := ids(c)
	if !ok {
		return
	}
	var body roleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "role is required")
		return
	}
	if err := ctrl.members.ChangeRole(c.Request.Context(), roomID, actorID, targetID, body.Role); err != nil {
		respondError(c, err, "Failed to change role")
		return
	}
	c.Status(http.StatusNoContent)
}

type reasonBody struct {
	Reason string `json:"reason" binding:"max=500"`
}

// Kick handles DELETE /moderation/rooms/:room_id/members/:user_id
func (ctrl *Controller) Kick(c *gin.Context) {
	actorID, roomID, targetID, ok := ids(c)
	if !ok {
		return
	}
	var body reasonBody
	_ = c.ShouldBindJSON(&body) // the reason is optional
	if err := ctrl.members.Kick(c.Request.Context(), roomID, actorID, targetID, body.Reason); err != nil {
		respondError(c, err, "Failed to remove member")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListAudit handles GET /moderation/rooms/:room_id/audit
func (ctrl *Controller) ListAudit(c *gin.Context) {
	actorID, roomID, _, ok := ids(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	entries, err := ctrl.audit.Execute(c.Request.Context(), roomID, actorID, limit)
	if err != nil {
		respondError(c, err, "Failed to load audit log")
		return
	}
	response.JSON(c, http.StatusOK, entries)
}

type reportBody struct {
	Reason string `json:"reason" binding:"required"`
}

// Report handles POST /moderation/messages/:message_id/report
func (ctrl *Controller) Report(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	messageID, err := uuid.Parse(c.Param("message_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid message ID format")
		return
	}
	var body reportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "reason is required")
		return
	}
	if _, err := ctrl.reports.Report(c.Request.Context(), messageID, userID, body.Reason); err != nil {
		respondError(c, err, "Failed to report message")
		return
	}
	c.Status(http.StatusCreated)
}

// ListReports handles GET /moderation/rooms/:room_id/reports?status=pending|resolved|ignored
func (ctrl *Controller) ListReports(c *gin.Context) {
	actorID, roomID, _, ok := ids(c)
	if !ok {
		return
	}
	reports, err := ctrl.reports.List(c.Request.Context(), roomID, actorID, c.Query("status"))
	if err != nil {
		respondError(c, err, "Failed to list reports")
		return
	}
	response.JSON(c, http.StatusOK, reports)
}

type resolveBody struct {
	Action        string `json:"action" binding:"required"` // "resolve" or "ignore"
	DeleteMessage bool   `json:"delete_message"`
}

// ResolveReport handles POST /moderation/reports/:report_id/resolve
func (ctrl *Controller) ResolveReport(c *gin.Context) {
	actorID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	reportID, err := strconv.Atoi(c.Param("report_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid report ID")
		return
	}
	var body resolveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "action is required")
		return
	}
	if err := ctrl.reports.Resolve(c.Request.Context(), reportID, actorID, body.Action, body.DeleteMessage); err != nil {
		respondError(c, err, "Failed to resolve report")
		return
	}
	c.Status(http.StatusNoContent)
}

type banBody struct {
	Reason   string `json:"reason" binding:"max=500"`
	Duration string `json:"duration"` // Go duration such as "24h"; empty = permanent
}

// Ban handles POST /moderation/rooms/:room_id/bans/:user_id
func (ctrl *Controller) Ban(c *gin.Context) {
	actorID, roomID, targetID, ok := ids(c)
	if !ok {
		return
	}
	var body banBody
	_ = c.ShouldBindJSON(&body)
	if err := ctrl.bans.Ban(c.Request.Context(), roomID, actorID, targetID, body.Reason, body.Duration); err != nil {
		respondError(c, err, "Failed to ban user")
		return
	}
	c.Status(http.StatusNoContent)
}

// Unban handles DELETE /moderation/rooms/:room_id/bans/:user_id
func (ctrl *Controller) Unban(c *gin.Context) {
	actorID, roomID, targetID, ok := ids(c)
	if !ok {
		return
	}
	if err := ctrl.bans.Unban(c.Request.Context(), roomID, actorID, targetID); err != nil {
		respondError(c, err, "Failed to lift ban")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListBans handles GET /moderation/rooms/:room_id/bans
func (ctrl *Controller) ListBans(c *gin.Context) {
	actorID, roomID, _, ok := ids(c)
	if !ok {
		return
	}
	bans, err := ctrl.bans.List(c.Request.Context(), roomID, actorID)
	if err != nil {
		respondError(c, err, "Failed to list bans")
		return
	}
	response.JSON(c, http.StatusOK, bans)
}
