package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_UpdateMany(t *testing.T) {
	opts := buildUpdateManyOpts([]UpdateManyOpt{WithUpsertMany(true)})
	assert.NotNil(t, opts)
}

func TestOptions_BulkWrite(t *testing.T) {
	opts := buildBulkWriteOpts([]BulkWriteOpt{WithOrdered(false)})
	assert.NotNil(t, opts)
}
