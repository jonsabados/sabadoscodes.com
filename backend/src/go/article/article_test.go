package article

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCachedLister(t *testing.T) {
	ctx := context.Background()

	summaries := map[bool]*struct {
		callCount int
		summaries []Summary
	}{
		true: {callCount: 0, summaries: []Summary{
			{
				Slug: "published",
			},
		}},
		false: {callCount: 0, summaries: []Summary{
			{
				Slug: "unpublished",
			},
		}},
	}

	base := Lister(func(ctx context.Context, published bool) ([]Summary, error) {
		summaries[published].callCount++
		return summaries[published].summaries, nil
	})

	cacheTime := time.Millisecond * 50

	testInstance := NewCachedLister(base, cacheTime)

	res, err := testInstance(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, summaries[true].summaries, res)

	res, err = testInstance(ctx, false)
	assert.NoError(t, err)
	assert.Equal(t, summaries[false].summaries, res)

	res, err = testInstance(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, summaries[true].summaries, res)
	assert.Equal(t, 1, summaries[true].callCount)

	res, err = testInstance(ctx, false)
	assert.NoError(t, err)
	assert.Equal(t, summaries[false].summaries, res)
	assert.Equal(t, 1, summaries[false].callCount)

	time.Sleep(cacheTime)

	res, err = testInstance(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, summaries[true].summaries, res)
	assert.Equal(t, 2, summaries[true].callCount)

	res, err = testInstance(ctx, false)
	assert.NoError(t, err)
	assert.Equal(t, summaries[false].summaries, res)
	assert.Equal(t, 2, summaries[false].callCount)
}
