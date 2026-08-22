package core

import "fmt"

// classificationVocab is the closed vocabulary that keeps Classification from
// degrading into EAV (design §3.3, guardrail 1: "Closed vocabulary. Each axis has a
// constrained set of allowed values … invalid combinations cannot be written."). In
// the full system this reference is governed/versioned; here it is a small in-code set.
var classificationVocab = map[EnumClassificationAxis]map[string]struct{}{
	EnumClassificationAxisSizeSegment: {"micro": {}, "sme": {}, "corporate": {}},
	EnumClassificationAxisServiceTier: {"gold": {}, "silver": {}, "bronze": {}},
}

// ValidateClassification enforces the closed vocabulary. An (axis, value) pair that is
// not in the vocabulary cannot be written — this is checked before persistence.
func ValidateClassification(axis EnumClassificationAxis, value string) error {
	allowed, ok := classificationVocab[axis]
	if !ok {
		return fmt.Errorf("unknown classification axis %q", axis)
	}
	if _, ok := allowed[value]; !ok {
		return fmt.Errorf("value %q is not in the closed vocabulary for axis %q", value, axis)
	}
	return nil
}
