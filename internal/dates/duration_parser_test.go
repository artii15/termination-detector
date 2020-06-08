package dates_test

import (
	"testing"
	"time"

	"github.com/artii15/termination-detector/internal/dates"
	"github.com/stretchr/testify/assert"
)

func TestMustParseDuration(t *testing.T) {
	assert.Equal(t, time.Hour*12, dates.MustParseDuration("12h"))
	assert.Panics(t, func() {
		dates.MustParseDuration("bad")
	})
}
