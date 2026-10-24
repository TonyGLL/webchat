package application

import (
	"context"
	"strings"

	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

const (
	minReasonLength = 3
	maxReasonLength = 500
	maxReportsList  = 100
)

// ReportsUseCase lets members report messages and moderators work through the reports.
type ReportsUseCase struct {
	repo     domain.Repository
	realtime domain.Realtime
}

func NewReportsUseCase(repo domain.Repository, realtime domain.Realtime) *ReportsUseCase {
	return &ReportsUseCase{repo: repo, realtime: realtime}
}

// Report files a complaint. Reporters must be members of the room, cannot report themselves,
// and can report each message only once.
func (uc *ReportsUseCase) Report(ctx context.Context, messageID uuid.UUID, reporterID int, reason string) (*domain.Report, error) {
	reason = strings.TrimSpace(reason)
	if n := len([]rune(reason)); n < minReasonLength || n > maxReasonLength {
		return nil, shared_domain.ErrInvalidInput
	}

	msg, err := uc.repo.MessageRef(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg.Deleted {
		return nil, domain.ErrMessageNotFound
	}
	actor, err := uc.repo.ActorIn(ctx, msg.RoomID, reporterID)
	if err != nil {
		return nil, err
	}
	if !actor.IsMember {
		return nil, shared_domain.ErrForbidden
	}
	if msg.Author != nil && *msg.Author == reporterID {
		return nil, shared_domain.ErrInvalidInput
	}
	return uc.repo.CreateReport(ctx, messageID, reporterID, reason)
}

// List returns a room's reports (default: pending) to moderators and above.
func (uc *ReportsUseCase) List(ctx context.Context, roomID uuid.UUID, userID int, status string) ([]*domain.Report, error) {
	actor, err := uc.repo.ActorIn(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !actor.CanResolveReports() {
		return nil, shared_domain.ErrForbidden
	}
	switch status {
	case "":
		status = domain.ReportPending
	case domain.ReportPending, domain.ReportResolved, domain.ReportIgnored:
	default:
		return nil, shared_domain.ErrInvalidInput
	}
	return uc.repo.ListReports(ctx, roomID, status, maxReportsList)
}

// Resolve closes a report: "resolve" (optionally deleting the message) or "ignore".
func (uc *ReportsUseCase) Resolve(ctx context.Context, reportID, actorID int, action string, deleteMessage bool) error {
	if action != "resolve" && action != "ignore" {
		return shared_domain.ErrInvalidInput
	}
	report, err := uc.repo.FindReport(ctx, reportID)
	if err != nil {
		return err
	}
	actor, err := uc.repo.ActorIn(ctx, report.RoomID, actorID)
	if err != nil {
		return err
	}
	if !actor.CanResolveReports() {
		return shared_domain.ErrForbidden
	}
	if report.Status != domain.ReportPending {
		return shared_domain.ErrConflict
	}

	status, auditAction := domain.ReportIgnored, domain.ActionReportIgnored
	if action == "resolve" {
		status, auditAction = domain.ReportResolved, domain.ActionReportResolved
	}

	deleted := false
	if action == "resolve" && deleteMessage {
		msg, err := uc.repo.MessageRef(ctx, report.MessageID)
		if err == nil && !msg.Deleted {
			if err := uc.repo.SoftDeleteMessage(ctx, report.MessageID); err != nil {
				return err
			}
			deleted = true
			if uc.realtime != nil {
				uc.realtime.Broadcast("delete_message", map[string]interface{}{"message_id": report.MessageID}, report.RoomID.String())
			}
		}
	}

	if err := uc.repo.SetReportStatus(ctx, reportID, status, actorID); err != nil {
		return err
	}
	roomID, messageID := report.RoomID, report.MessageID
	if err := uc.repo.AddAudit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: auditAction, RoomID: &roomID, TargetUserID: report.MessageAuthor, TargetMessageID: &messageID,
		Details: map[string]interface{}{"report_id": reportID, "message_deleted": deleted},
	}); err != nil {
		logAuditFailure(auditAction, err)
	}
	return nil
}
