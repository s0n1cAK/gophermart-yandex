package jobs

import (
	"context"
	"testing"
	"yandex-diplom/internal/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBalanceJob(t *testing.T) {
	j := BalanceJob{}
	mockSvc := new(mocks.MockService)

	mockSvc.On("UpdateMissingBalanceEntries", mock.Anything).Return(nil)

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
	mockSvc.AssertCalled(t, "UpdateMissingBalanceEntries", mock.Anything)
}
