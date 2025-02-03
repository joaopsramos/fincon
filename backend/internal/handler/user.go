package handler

import (
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v3"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	userRepo domain.UserRepo
}

var userCreateSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Max(160).Email().Required(),
	"password": z.String().Trim().Min(8).Max(72).Required(),
})

var userLoginSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Required(),
	"password": z.String().Trim().Required(),
})

func NewUserHandler(userRepo domain.UserRepo) UserHandler {
	return UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) Create(ctx fiber.Ctx) error {
	var params struct {
		Email    string
		Password string
	}

	if errs := util.ParseZodSchema(userCreateSchema, ctx.Body(), &params); errs != nil {
		return ctx.Status(http.StatusBadRequest).JSON(util.M{"errors": errs})
	}

	if _, err := h.userRepo.GetByEmail(params.Email); err == nil {
		return ctx.Status(http.StatusConflict).JSON(util.M{"error": "email already in use"})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := &domain.User{Email: params.Email, HashPassword: string(hashPassword)}
	err = h.userRepo.Create(user)
	if err != nil {
		panic(err)
	}

	token, err := user.CreateToken()
	if err != nil {
		panic(err)
	}

	return ctx.Status(http.StatusCreated).JSON(util.M{"user": user, "token": token})
}

func (h *UserHandler) Login(ctx fiber.Ctx) error {
	var params struct {
		Email    string
		Password string
	}

	if errs := util.ParseZodSchema(userLoginSchema, ctx.Body(), &params); errs != nil {
		return ctx.Status(http.StatusBadRequest).JSON(util.M{"errors": errs})
	}

	user, err := h.userRepo.GetByEmail(params.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.M{"error": "invalid email or password"})
	} else if err != nil {
		panic(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(params.Password)); err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.M{"error": "invalid email or password"})
	}

	token, err := user.CreateToken()
	if err != nil {
		panic(err)
	}

	return ctx.Status(http.StatusCreated).JSON(util.M{"token": token})
}
