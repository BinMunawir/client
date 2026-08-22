package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/BinMunawir/client/internal/client/store"
	"github.com/BinMunawir/client/internal/core"
	"github.com/go-jet/jet/v2/qrm"
)

type ActivateInput struct {
	Biz core.Business
}

// Activate transitions the Business draft → active (design §7). Skeleton: persist →
// classify (standard §6.2).
func Activate(ctx context.Context, in ActivateInput) (core.Business, error) {
	biz, err := activatePersist(ctx, in.Biz)
	if err != nil {
		return in.Biz, fmt.Errorf("activatePersist: %w", err)
	}
	return biz, nil
}

func activatePersist(ctx context.Context, biz core.Business) (core.Business, error) {
	updated, err := store.ActivateBusiness(ctx, biz)
	if err == nil {
		return updated, nil
	}
	if errors.Is(err, qrm.ErrNoRows) {
		// The guarded transition matched no row: the business is already active (an idempotent
		// replay). Re-read and return it as a no-op success (standard §6.2 classify + §8 no-op).
		current, ferr := store.FetchBusiness(ctx, biz.ID)
		if ferr != nil {
			return biz, fmt.Errorf("activatePersist: refetch after guarded no-op: %w", ferr)
		}
		if current.Status == core.EnumBusinessStatusActive {
			return current, nil
		}
		return biz, fmt.Errorf("activatePersist: guarded activate matched no row, status is %s: %w", current.Status, err)
	}
	return biz, fmt.Errorf("store.ActivateBusiness: %w", err)
}
