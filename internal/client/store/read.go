package store

import (
	"context"
	"fmt"

	"github.com/BinMunawir/client/internal/adapters/jet"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/table"
	"github.com/BinMunawir/client/internal/core"
	"github.com/go-jet/jet/v2/postgres"
)

func FetchBusiness(ctx context.Context, id string) (core.Business, error) {
	stmt := table.Businesses.
		SELECT(table.Businesses.AllColumns).
		WHERE(table.Businesses.ID.EQ(postgres.String(id)))

	var m model.Businesses
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Business{}, fmt.Errorf("pg: fetch business %q: %w", id, err)
	}
	return BusinessFromModel(m), nil
}

// FetchBusinessByCorrID resolves a business by its caller-supplied correlation key. Used
// by the idempotent create path to reconcile a unique-violation replay (standard §8).
func FetchBusinessByCorrID(ctx context.Context, corrID string) (core.Business, error) {
	stmt := table.Businesses.
		SELECT(table.Businesses.AllColumns).
		WHERE(table.Businesses.CorrelationID.EQ(postgres.String(corrID)))

	var m model.Businesses
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return core.Business{}, fmt.Errorf("pg: fetch business by corr_id %q: %w", corrID, err)
	}
	return BusinessFromModel(m), nil
}

// FetchAllBusinesses returns a page of businesses. The base statement is built once and the
// count is derived from it — no second hand-built query (standard §6.3).
func FetchAllBusinesses(ctx context.Context, limit, offset int) (pg.Pagination[core.Business], error) {
	stmt := postgres.SELECT(table.Businesses.AllColumns).FROM(table.Businesses)

	page := pg.Pagination[core.Business]{}
	stmtCount := postgres.SELECT(postgres.COUNT(postgres.STAR).AS("total")).
		FROM(stmt.AsTable(table.Businesses.TableName()))
	if err := stmtCount.QueryContext(ctx, pg.DB(), &page); err != nil {
		return pg.Pagination[core.Business]{}, fmt.Errorf("pg: count businesses: %w", err)
	}
	page.Limit, page.Offset = page.NormalizePaging(limit, offset)

	if page.Total == 0 || int64(offset) >= page.Total {
		return page, nil
	}

	stmtRows := stmt.
		ORDER_BY(table.Businesses.CreatedAt.DESC()).
		LIMIT(int64(page.Limit)).
		OFFSET(int64(page.Offset))

	var mdls []model.Businesses
	if err := stmtRows.QueryContext(ctx, pg.DB(), &mdls); err != nil {
		return pg.Pagination[core.Business]{}, fmt.Errorf("pg: list businesses: %w", err)
	}

	page.Items = make([]core.Business, len(mdls))
	for i, m := range mdls {
		page.Items[i] = BusinessFromModel(m)
	}
	return page, nil
}

// FetchClassifications returns the labeled (axis, value) records attached to a business.
func FetchClassifications(ctx context.Context, businessID string) ([]core.Classification, error) {
	stmt := table.Classifications.
		SELECT(table.Classifications.AllColumns).
		WHERE(table.Classifications.BusinessID.EQ(postgres.String(businessID))).
		ORDER_BY(table.Classifications.Axis.ASC())

	var mdls []model.Classifications
	if err := stmt.QueryContext(ctx, pg.DB(), &mdls); err != nil {
		return nil, fmt.Errorf("pg: list classifications for %q: %w", businessID, err)
	}
	out := make([]core.Classification, len(mdls))
	for i, m := range mdls {
		out[i] = ClassificationFromModel(m)
	}
	return out, nil
}
