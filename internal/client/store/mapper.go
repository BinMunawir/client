package store

import (
	"github.com/BinMunawir/client/internal/adapters/jet/.gen/clientdb/public/model"
	"github.com/BinMunawir/client/internal/core"
)

// Mappers convert between domain entities and the generated DB models. Unlike a value
// domain, this domain carries no money, so there are no decimals to convert here — the
// mapping is a straight field copy (nullable columns are pointers on both sides).

func BusinessToModel(b core.Business) model.Businesses {
	return model.Businesses{
		ID:                b.ID,
		CorrelationID:     b.CorrID,
		LegalName:         b.LegalName,
		TradeName:         b.TradeName,
		CrNumber:          b.CRNumber,
		LegalForm:         string(b.LegalForm),
		IncorporationDate: b.IncorporationDate,
		Status:            string(b.Status),
		OrganizationID:    b.OrganizationID,
		KeycloakOrgID:     b.KeycloakOrgID,
		LedgerRef:         b.LedgerRef,
		UpdatedAt:         b.UpdatedAt,
		CreatedAt:         b.CreatedAt,
	}
}

func BusinessFromModel(m model.Businesses) core.Business {
	return core.Business{
		ID:                m.ID,
		CorrID:            m.CorrelationID,
		LegalName:         m.LegalName,
		TradeName:         m.TradeName,
		CRNumber:          m.CrNumber,
		LegalForm:         core.EnumLegalForm(m.LegalForm),
		IncorporationDate: m.IncorporationDate,
		Status:            core.EnumBusinessStatus(m.Status),
		OrganizationID:    m.OrganizationID,
		KeycloakOrgID:     m.KeycloakOrgID,
		LedgerRef:         m.LedgerRef,
		UpdatedAt:         m.UpdatedAt,
		CreatedAt:         m.CreatedAt,
	}
}

func OrganizationToModel(o core.Organization) model.Organizations {
	return model.Organizations{
		ID:        o.ID,
		Name:      o.Name,
		UpdatedAt: o.UpdatedAt,
		CreatedAt: o.CreatedAt,
	}
}

func OrganizationFromModel(m model.Organizations) core.Organization {
	return core.Organization{
		ID:        m.ID,
		Name:      m.Name,
		UpdatedAt: m.UpdatedAt,
		CreatedAt: m.CreatedAt,
	}
}

func ClassificationToModel(c core.Classification) model.Classifications {
	return model.Classifications{
		ID:         c.ID,
		BusinessID: c.BusinessID,
		Axis:       string(c.Axis),
		Value:      c.Value,
		UpdatedAt:  c.UpdatedAt,
		CreatedAt:  c.CreatedAt,
	}
}

func ClassificationFromModel(m model.Classifications) core.Classification {
	return core.Classification{
		ID:         m.ID,
		BusinessID: m.BusinessID,
		Axis:       core.EnumClassificationAxis(m.Axis),
		Value:      m.Value,
		UpdatedAt:  m.UpdatedAt,
		CreatedAt:  m.CreatedAt,
	}
}
