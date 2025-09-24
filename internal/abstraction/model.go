package abstraction

import (
	"time"
)

type EntityJustCreated struct {
	CreatedAt time.Time `json:"created_at"`
}

type Entity struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type EntityWithBy struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedBy int        `json:"created_by"`
	UpdatedBy *int       `json:"updated_by"`
}
