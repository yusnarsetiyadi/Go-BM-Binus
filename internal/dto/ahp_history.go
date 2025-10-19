package dto

type AhpHistoryCreateRequest struct {
	Kriteria             []string                          `json:"kriteria" validate:"required"`
	KriteriaComparison   []AhpComparisonRequest            `json:"kriteria_comparison" validate:"required"`
	Alternatif           []string                          `json:"alternatif" validate:"required"`
	AlternatifComparison map[string][]AhpComparisonRequest `json:"alternatif_comparison" validate:"required"`
	ReferenceRequest     int                               `json:"reference_request" validate:"required"`
}

type AhpComparisonRequest struct {
	Item1 string  `json:"item1"`
	Item2 string  `json:"item2"`
	Value float64 `json:"value"`
}

type AhpHistoryFindByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type AhpHistoryDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}
