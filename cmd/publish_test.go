package cmd

import (
	"testing"
)

func TestPublish(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	client.On("PublishTheme").Return(nil)
	publish(ctx)
	client.AssertExpectations(t)
}
