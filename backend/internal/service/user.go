package service

import (
	"github.com/joaopsramos/fincon/internal/domain"
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
