package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/mail"
	"github.com/joaopsramos/fincon/internal/types"
	"github.com/joaopsramos/fincon/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo domain.UserRepo
	mailer   mail.Mailer
}

type CreateUserDTO struct {
	Email    string
	Password string

	CreateSalaryDTO
}

type ResetPasswordDTO struct {
	Token    string
	Password string
}

func NewUserService(userRepo domain.UserRepo, mailer mail.Mailer) UserService {
	return UserService{userRepo: userRepo, mailer: mailer}
}

func (s *UserService) SendForgotPasswordEmail(ctx context.Context, user domain.User) error {
	userToken := &domain.UserToken{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		Used:      false,
	}

	if err := s.userRepo.CreateToken(ctx, userToken); err != nil {
		return err
	}

	email := mail.Email{
		To:       types.MailContact{Email: user.Email},
		Subject:  mail.ForgotPasswordSubject,
		Template: mail.ForgotPasswordTemplate,
		Data:     util.M{"Link": fmt.Sprintf("%s/password/reset?token=%s", config.Get().WebURL, userToken.Token)},
	}

	return s.mailer.Send(email)
}

func (s *UserService) ResetPassword(ctx context.Context, dto ResetPasswordDTO) error {
	if err := uuid.Validate(dto.Token); err != nil {
		return errs.ErrInvalidToken
	}

	token, err := s.userRepo.GetUserTokenByToken(ctx, dto.Token)
	if err != nil {
		return err
	}

	if token.Used || time.Now().UTC().After(token.ExpiresAt) {
		return errs.ErrInvalidToken
	}

	hashPassword, err := s.generatePassword([]byte(dto.Password))
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdateUserPassword(ctx, token.UserID, string(hashPassword)); err != nil {
		return err
	}

	return s.userRepo.MarkTokenAsUsed(ctx, token.ID)
}

func (s *UserService) Create(ctx context.Context, dto CreateUserDTO) (*domain.User, *domain.Salary, error) {
	hashPassword, err := s.generatePassword([]byte(dto.Password))
	if err != nil {
		panic(err)
	}

	user := domain.User{
		Email:        dto.Email,
		HashPassword: string(hashPassword),
	}

	salary := BuildSalary(dto.CreateSalaryDTO)

	if err := s.userRepo.Create(ctx, &user, &salary); err != nil {
		return &domain.User{}, &domain.Salary{}, err
	}

	return &user, &salary, nil
}

func (s *UserService) GetByEmailAndPassword(ctx context.Context, email string, password string) (*domain.User, error) {
	user, err := s.GetByEmail(ctx, email)
	if errors.Is(err, errs.ErrNotFound{}) {
		return &domain.User{}, errors.Join(err, errs.ErrInvalidCredentials)
	} else if err != nil {
		return &domain.User{}, err
	}

	if !s.isSamePassword(*user, password) {
		return &domain.User{}, errs.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *UserService) isSamePassword(user domain.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password))
	return err == nil
}

func (s *UserService) generatePassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}
