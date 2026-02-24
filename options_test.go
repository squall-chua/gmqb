package gmqb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_UpdateMany(t *testing.T) {
	opts := buildUpdateManyOpts([]UpdateManyOpt{WithUpsertMany(true)})
	assert.NotNil(t, opts)
}
