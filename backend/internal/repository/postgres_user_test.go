package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestPostgresUser_Create(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	repo := repository.NewPostgresUser(tx)

	user := domain.User{Email: "test@mail.com", HashPassword: "pass"}
	salary := domain.Salary{Amount: 1000}
	assert.NoError(repo.Create(context.Background(), &user, &salary))

	assert.NotZero(user.ID)
	assert.NotZero(salary.ID)

	goals := []map[string]any{}
	tx.Model(domain.Goal{}).Where("user_id =?", user.ID).Select("name, percentage").Scan(&goals)

	assert.ElementsMatch(goals, []map[string]any{
		{"name": "Fixed costs", "percentage": int64(40)},
		{"name": "Comfort", "percentage": int64(20)},
		{"name": "Goals", "percentage": int64(5)},
		{"name": "Pleasures", "percentage": int64(5)},
		{"name": "Financial investments", "percentage": int64(25)},
		{"name": "Knowledge", "percentage": int64(5)},
	})
}

func TestPostgresUser_CreateToken(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	repo := repository.NewPostgresUser(tx)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()

	token := domain.UserToken{UserID: user.ID}
	assert.NoError(repo.CreateToken(context.Background(), &token))
	assert.NotZero(token.ID)
	assert.Equal(user.ID, token.UserID)
}

func TestPostgresUser_GetUserTokenByToken(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	factory := testhelper.NewFactory(tx)

	tests := []struct {
		name        string
		setupToken  func(domain.User) domain.UserToken
		expectedErr error
	}{
		{
			name: "success with valid user token",
			setupToken: func(user domain.User) domain.UserToken {
				return factory.InsertUserToken(&domain.UserToken{UserID: user.ID})
			},
		},
		{
			name: "token not found",
			setupToken: func(user domain.User) domain.UserToken {
				// insert a token to make sure the database is not empty
				factory.InsertUserToken(&domain.UserToken{UserID: user.ID})
				return domain.UserToken{Token: uuid.New()}
			},
			expectedErr: errs.NewNotFound("token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			token := tt.setupToken(factory.InsertUser())

			repo := repository.NewPostgresUser(tx)
			gotToken, err := repo.GetUserTokenByToken(context.Background(), token.Token.String())

			if tt.expectedErr != nil {
				assert.Error(err)
				assert.Equal(tt.expectedErr, err)
				return
			}

			assert.NoError(err)
			assert.Equal(token.ID, gotToken.ID)
			assert.Equal(token.UserID, gotToken.UserID)
			assert.Equal(token.Token, gotToken.Token)
			assert.Equal(token.Used, gotToken.Used)
			assert.WithinDuration(token.ExpiresAt, gotToken.ExpiresAt, time.Second)
		})
	}
}

func TestPostgresUser_UpdateUserPassword(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	factory := testhelper.NewFactory(tx)

	tests := []struct {
		name        string
		setupUser   func() (uuid.UUID, string)
		expectedErr error
	}{
		{
			name: "success updating password",
			setupUser: func() (uuid.UUID, string) {
				user := factory.InsertUser()
				return user.ID, "new-hashed-password"
			},
		},
		{
			name: "not found when user does not exist",
			setupUser: func() (uuid.UUID, string) {
				return uuid.New(), "new-hashed-password"
			},
			expectedErr: errs.NewNotFound("user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			userID, newPassword := tt.setupUser()

			repo := repository.NewPostgresUser(tx)
			err := repo.UpdateUserPassword(context.Background(), userID, newPassword)

			if tt.expectedErr != nil {
				assert.Error(err)
				assert.Equal(tt.expectedErr, err)
				return
			}

			assert.NoError(err)

			var updatedUser domain.User
			tx.First(&updatedUser, userID)
			assert.Equal(newPassword, updatedUser.HashPassword)
		})
	}
}

func TestPostgresUser_MarkTokenAsUsed(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	factory := testhelper.NewFactory(tx)

	tests := []struct {
		name        string
		setupToken  func() uint
		expectedErr error
	}{
		{
			name: "success marking token as used",
			setupToken: func() uint {
				user := factory.InsertUser()
				token := factory.InsertUserToken(&domain.UserToken{UserID: user.ID})
				return token.ID
			},
		},
		{
			name: "not found when token does not exist",
			setupToken: func() uint {
				return uint(9999)
			},
			expectedErr: errs.NewNotFound("token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			tokenID := tt.setupToken()

			repo := repository.NewPostgresUser(tx)
			err := repo.MarkTokenAsUsed(context.Background(), tokenID)

			if tt.expectedErr != nil {
				assert.Error(err)
				assert.Equal(tt.expectedErr, err)
				return
			}

			assert.NoError(err)

			// Verify the token was actually marked as used
			var updatedToken domain.UserToken
			tx.First(&updatedToken, tokenID)
			assert.True(updatedToken.Used)
		})
	}
}
