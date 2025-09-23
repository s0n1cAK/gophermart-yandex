package gophermart

import (
	"testing"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/require"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    models.User
		wantErr error
	}{
		{
			name:    "empty login and password",
			user:    models.User{Login: "", Password: ""},
			wantErr: domain.ErrInvalidPayload,
		},
		{
			name:    "short login",
			user:    models.User{Login: "ab", Password: "strongpass"},
			wantErr: domain.ErrInvalidPayload,
		},
		{
			name:    "short password",
			user:    models.User{Login: "validlogin", Password: "pw"},
			wantErr: domain.ErrInvalidPayload,
		},
		{
			name:    "valid user",
			user:    models.User{Login: "validlogin", Password: "validpass"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, domain.GetAppErr(err) != nil)
				require.True(t, domain.GetAppErr(err).Error() != "")
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
