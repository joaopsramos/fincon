package service

import (
	"errors"

	"github.com/joaopsramos/fincon/internal/domain"
	errs "github.com/joaopsramos/fincon/internal/error"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo domain.UserRepo
}

type CreateUserDTO struct {
	Email    string
	Password string

	CreateSalaryDTO
}

func NewUserService(userRepo domain.UserRepo) UserService {
	return UserService{userRepo: userRepo}
}

func (s *UserService) Create(dto CreateUserDTO) (*domain.User, *domain.Salary, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := domain.User{
		Email:        dto.Email,
		HashPassword: string(hashPassword),
	}

	salary := NewSalary(dto.CreateSalaryDTO)

	if err := s.userRepo.Create(&user, &salary); err != nil {
		return &domain.User{}, &domain.Salary{}, err
	}

	return &user, &salary, nil
}

func (s *UserService) GetByEmailAndPassword(email string, password string) (*domain.User, error) {
	user, err := s.GetByEmail(email)
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

func (s *UserService) GetByEmail(email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(email)
}

func (s *UserService) isSamePassword(user domain.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password))
	return err == nil
}
