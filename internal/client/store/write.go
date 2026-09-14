package store

import (
	"context"
	"fmt"
	"time"

	"github.com/BinMunawir/client/internal/adapters/jet"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/table"
	"github.com/BinMunawir/client/internal/core"
	"github.com/go-jet/jet/v2/postgres"
)

func InsertBusiness(ctx context.Context, b core.Business) (core.Business, error) {
	stmt := table.Businesses.
		INSERT(table.Businesses.AllColumns).
		MODEL(BusinessToModel(b)).
		RETURNING(table.Businesses.AllColumns)

	var m model.Businesses
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return b, fmt.Errorf("stmt.QueryContext %q: %w", b.ID, err)
	}
	return BusinessFromModel(m), nil
}

func InsertOrganization(ctx context.Context, o core.Organization) (core.Organization, error) {
	stmt := table.Organizations.
		INSERT(table.Organizations.AllColumns).
		MODEL(OrganizationToModel(o)).
		RETURNING(table.Organizations.AllColumns)

	var m model.Organizations
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return o, fmt.Errorf("stmt.QueryContext %q: %w", o.ID, err)
	}
	return OrganizationFromModel(m), nil
}

func InsertClassification(ctx context.Context, c core.Classification) (core.Classification, error) {
	stmt := table.Classifications.
		INSERT(table.Classifications.AllColumns).
		MODEL(ClassificationToModel(c)).
		RETURNING(table.Classifications.AllColumns)

	var m model.Classifications
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return c, fmt.Errorf("stmt.QueryContext %q: %w", c.ID, err)
	}
	return ClassificationFromModel(m), nil
}

// SetBusinessKeycloakOrg records the reference-out to the Keycloak Organization. It is an
// unconditional set keyed on the id (setting the same value again is a harmless no-op), so
// a replayed provisioning step converges rather than erroring.
func SetBusinessKeycloakOrg(ctx context.Context, b core.Business) (core.Business, error) {
	stmt := table.Businesses.UPDATE().SET(
		table.Businesses.KeycloakOrgID.SET(postgres.String(deref(b.KeycloakOrgID))),
		table.Businesses.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	).WHERE(
		table.Businesses.ID.EQ(postgres.String(b.ID)),
	).RETURNING(table.Businesses.AllColumns)

	var m model.Businesses
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return b, fmt.Errorf("QueryContext: %w", err)
	}
	return BusinessFromModel(m), nil
}

// Activate / Suspend / Offboard are guarded status transitions (design §7, standard §6.3).

func ActivateBusiness(ctx context.Context, b core.Business) (core.Business, error) {
	return updateBusinessStatus(ctx, b, core.EnumBusinessStatusActive)
}

func SuspendBusiness(ctx context.Context, b core.Business) (core.Business, error) {
	return updateBusinessStatus(ctx, b, core.EnumBusinessStatusSuspended)
}

func OffboardBusiness(ctx context.Context, b core.Business) (core.Business, error) {
	return updateBusinessStatus(ctx, b, core.EnumBusinessStatusOffboarded)
}

// updateBusinessStatus is the single guarded status-transition helper. The
// `WHERE status <> target` guard makes each transition idempotent; RETURNING re-reads the
// row so the caller gets canonical state. A no-match surfaces as qrm.ErrNoRows and is
// classified by the caller (standard §6.2).
func updateBusinessStatus(ctx context.Context, b core.Business, to core.EnumBusinessStatus, extras ...postgres.ColumnAssigment) (core.Business, error) {
	sets := []any{
		table.Businesses.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	}
	for _, extra := range extras {
		sets = append(sets, extra)
	}

	stmt := table.Businesses.UPDATE().SET(
		table.Businesses.Status.SET(postgres.String(string(to))),
		sets...,
	).WHERE(
		table.Businesses.ID.EQ(postgres.String(b.ID)).
			AND(table.Businesses.Status.NOT_EQ(postgres.String(string(to)))),
	).RETURNING(table.Businesses.AllColumns)

	var m model.Businesses
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return b, fmt.Errorf("QueryContext: %w", err)
	}
	return BusinessFromModel(m), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
