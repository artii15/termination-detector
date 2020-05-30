package dates_test

import (
	"testing"
	"time"

	"github.com/nordcloud/termination-detector/internal/dates"
	"github.com/stretchr/testify/assert"
)

func TestCurrentDateGetter_GetCurrentDate_IsUTC(t *testing.T) {
	getter := dates.NewCurrentDateGetter()
	date := getter.GetCurrentDate()
	assert.Equal(t, time.UTC, date.Location())
}
