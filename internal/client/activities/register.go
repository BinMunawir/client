package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BinMunawir/client/internal/client/store"
	"github.com/BinMunawir/client/internal/core"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// RegisterInput is the flat, serializable capture of a registration. LegalForm branches
// which fields apply at capture (design §3.3) but the result is one Business entity.
type RegisterInput struct {
	CorrID            string
	LegalName         string
	TradeName         string
	CRNumber          string
	LegalForm         string
	IncorporationDate *time.Time
	OrganizationID    *string
	SizeSegment       string // Classification value on axis SizeSegment (closed vocab)
	ServiceTier       string // Classification value on axis ServiceTier (closed vocab)
}

// Register creates the Business in `draft` and attaches its classifications. Skeleton:
// pure input→entity mapping → persist → classify errors (standard §6.2).
func Register(ctx context.Context, in RegisterInput) (core.Business, error) {
	biz, err := registerInputToEntity(in)
	if err != nil {
		return core.Business{}, fmt.Errorf("registerInputToEntity: %w", err)
	}
	biz, err = registerPersist(ctx, in, biz)
	if err != nil {
		return core.Business{}, fmt.Errorf("registerPersist: %w", err)
	}
	return biz, nil
}

func registerInputToEntity(in RegisterInput) (core.Business, error) {
	form := core.EnumLegalForm(in.LegalForm)
	switch form {
	case core.EnumLegalFormCompany, core.EnumLegalFormSoleEstablishment, core.EnumLegalFormFreelancer:
	default:
		return core.Business{}, fmt.Errorf("invalid legal_form %q", in.LegalForm)
	}

	// Enforce the closed classification vocabulary before anything is written (design §3.3).
	if in.SizeSegment != "" {
		if err := core.ValidateClassification(core.EnumClassificationAxisSizeSegment, in.SizeSegment); err != nil {
			return core.Business{}, err
		}
	}
	if in.ServiceTier != "" {
		if err := core.ValidateClassification(core.EnumClassificationAxisServiceTier, in.ServiceTier); err != nil {
			return core.Business{}, err
		}
	}

	now := time.Now().UTC()
	return core.Business{
		ID:                core.BusinessID(),
		CorrID:            in.CorrID,
		LegalName:         in.LegalName,
		TradeName:         in.TradeName,
		CRNumber:          in.CRNumber,
		LegalForm:         form,
		IncorporationDate: in.IncorporationDate,
		Status:            core.EnumBusinessStatusDraft,
		OrganizationID:    in.OrganizationID,
		UpdatedAt:         now,
		CreatedAt:         now,
	}, nil
}

func registerPersist(ctx context.Context, in RegisterInput, biz core.Business) (core.Business, error) {
	inserted, err := store.InsertBusiness(ctx, biz)
	if err != nil {
		// Idempotent create: a unique-violation on the correlation key means this run was
		// already registered — reconcile against the existing row (standard §8).
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return registerReconcile(ctx, biz, pgErr)
		}
		return biz, fmt.Errorf("store.InsertBusiness: %w", err)
	}

	if err := attachClassification(ctx, inserted.ID, core.EnumClassificationAxisSizeSegment, in.SizeSegment); err != nil {
		return inserted, err
	}
	if err := attachClassification(ctx, inserted.ID, core.EnumClassificationAxisServiceTier, in.ServiceTier); err != nil {
		return inserted, err
	}
	return inserted, nil
}

func registerReconcile(ctx context.Context, biz core.Business, pgErr *pgconn.PgError) (core.Business, error) {
	existing, err := store.FetchBusinessByCorrID(ctx, biz.CorrID)
	if err != nil {
		return biz, fmt.Errorf("registerReconcile: store.FetchBusinessByCorrID: %w", err)
	}
	if existing.CorrID != biz.CorrID {
		return biz, fmt.Errorf("registerReconcile: existing business has different correlation_id %s: %w", existing.CorrID, pgErr)
	}
	return existing, nil
}

// attachClassification writes one labeled (axis, value) record, tolerating a
// unique-violation on (business_id, axis) so a replay is idempotent.
func attachClassification(ctx context.Context, businessID string, axis core.EnumClassificationAxis, value string) error {
	if value == "" {
		return nil
	}
	now := time.Now().UTC()
	c := core.Classification{
		ID:         core.ClassificationID(),
		BusinessID: businessID,
		Axis:       axis,
		Value:      value,
		UpdatedAt:  now,
		CreatedAt:  now,
	}
	if _, err := store.InsertClassification(ctx, c); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil // already classified on this axis — idempotent
		}
		return fmt.Errorf("store.InsertClassification: %w", err)
	}
	return nil
}
