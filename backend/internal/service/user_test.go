package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/mail"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserService_SendForgotPasswordEmail(t *testing.T) {
	t.Parallel()
	mailer := testhelper.NewMockMailer(t)
	mockRepo := testhelper.NewMockUserRepo(t)

	tests := []struct {
		name       string
		repo       domain.UserRepo
		setupMocks func()
		postAssert func(t *testing.T, tx *gorm.DB, user domain.User)
		wantErr    bool
	}{
		{
			name: "create valid token",
			setupMocks: func() {
				mailer.EXPECT().Send(mock.AnythingOfType("mail.Email")).Return(nil).Once()
			},
			postAssert: func(t *testing.T, tx *gorm.DB, user domain.User) {
				assert := assert.New(t)

				var token domain.UserToken
				err := tx.First(&token, domain.UserToken{UserID: user.ID}).Error
				assert.NoError(err)
				assert.False(token.Used)
				assert.WithinDuration(time.Now().UTC().Add(24*time.Hour), token.ExpiresAt, time.Second*5)
			},
		},
		{
			name: "send email with created token",
			setupMocks: func() {
				mailer.EXPECT().Send(mock.AnythingOfType("mail.Email")).Return(nil).Once()
			},
			postAssert: func(t *testing.T, tx *gorm.DB, user domain.User) {
				assert := assert.New(t)

				var token domain.UserToken
				err := tx.First(&token, domain.UserToken{UserID: user.ID}).Error
				assert.NoError(err)

				mailer.AssertCalled(t, "Send", mock.MatchedBy(func(e mail.Email) bool {
					link := fmt.Sprintf("%s/password/reset?token=%s", config.Get().WebURL, token.Token)

					return e.To.Email == user.Email &&
						e.From.Email == "" &&
						e.Subject == mail.ForgotPasswordSubject &&
						e.Template == "forgot_password" &&
						e.Data["Link"].(string) == link
				}))
			},
		},
		// Use mocked repo to simulate error
		{
			name: "do not send email if token creation fails",
			repo: mockRepo,
			setupMocks: func() {
				mockRepo.EXPECT().CreateToken(mock.Anything, mock.Anything).Return(errors.New("error")).Once()
			},
			postAssert: func(t *testing.T, tx *gorm.DB, user domain.User) {
				mailer.AssertNotCalled(t, "Send")
			},
			wantErr: true,
		},
		{
			name: "return error if email sending fails",
			setupMocks: func() {
				mailer.EXPECT().Send(mock.AnythingOfType("mail.Email")).Return(errors.New("error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			tx := testhelper.NewTestPostgresTx(t)
			f := testhelper.NewFactory(tx)
			user := f.InsertUser()

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			if tt.repo == nil {
				tt.repo = repository.NewPostgresUser(tx)
			}

			s := service.NewUserService(tt.repo, mailer)
			err := s.SendForgotPasswordEmail(context.Background(), user)

			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tt.postAssert != nil {
				tt.postAssert(t, tx, user)
			}
		})
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	repo := repository.NewPostgresUser(tx)
	factory := testhelper.NewFactory(tx)

	tests := []struct {
		name        string
		setupToken  func(user domain.User) (uint, string)
		expectedErr error
	}{
		{
			name: "reset password with valid token",
			setupToken: func(user domain.User) (uint, string) {
				token := factory.InsertUserToken(&domain.UserToken{UserID: user.ID})
				return token.ID, token.Token.String()
			},
		},
		{
			name: "fail with expired token",
			setupToken: func(user domain.User) (uint, string) {
				token := factory.InsertUserToken(&domain.UserToken{
					UserID:    user.ID,
					ExpiresAt: time.Now().UTC().Add(-1 * time.Minute),
				})
				return token.ID, token.Token.String()
			},
			expectedErr: errs.ErrInvalidToken,
		},
		{
			name: "fail with used token",
			setupToken: func(user domain.User) (uint, string) {
				token := factory.InsertUserToken(&domain.UserToken{
					UserID: user.ID,
					Used:   true,
				})
				return token.ID, token.Token.String()
			},
			expectedErr: errs.ErrInvalidToken,
		},
		{
			name: "fail with token not found",
			setupToken: func(user domain.User) (uint, string) {
				return 0, uuid.NewString()
			},
			expectedErr: errs.NewNotFound("token"),
		},
		{
			name: "fail with invalid token",
			setupToken: func(user domain.User) (uint, string) {
				return 0, "invalid-token"
			},
			expectedErr: errs.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			user := factory.InsertUser()

			tokenID, token := tt.setupToken(user)
			newPassword := "new-password"

			s := service.NewUserService(repo, nil)
			err := s.ResetPassword(context.Background(), service.ResetPasswordDTO{
				Token:    token,
				Password: newPassword,
			})

			if tt.expectedErr != nil {
				assert.ErrorIs(err, tt.expectedErr)
				return
			}

			assert.NoError(err)

			var updatedUser domain.User
			tx.First(&updatedUser, user.ID)
			err = bcrypt.CompareHashAndPassword([]byte(updatedUser.HashPassword), []byte(newPassword))
			assert.NotEqual(updatedUser.HashPassword, user.HashPassword)
			assert.NoError(err)

			var userToken domain.UserToken
			tx.First(&userToken, tokenID)
			assert.True(userToken.Used)
		})
	}
}
