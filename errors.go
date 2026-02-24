package gmqb

import "errors"

var (
	// ErrInvalidField is returned when a struct field path cannot be resolved
	// to a BSON field name via struct tag reflection.
	ErrInvalidField = errors.New("gmqb: invalid field path")

	// ErrEmptyFilter is returned when an empty filter is passed where one is required.
	ErrEmptyFilter = errors.New("gmqb: empty filter")

	// ErrEmptyUpdate is returned when an empty update document is passed.
	ErrEmptyUpdate = errors.New("gmqb: empty update document")

	// ErrEmptyPipeline is returned when an empty pipeline is passed.
	ErrEmptyPipeline = errors.New("gmqb: empty pipeline")
)
