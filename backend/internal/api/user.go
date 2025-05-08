package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/auth"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

type UserIDKeyType string

var (
	tokenExpiresIn               = time.Hour * 24 * 7
	UserIDKey      UserIDKeyType = "user_id"
)

type UserHandler struct {
	*BaseHandler
	userService service.UserService
}

var (
	passwordFieldSchema = z.String().Trim().
				Min(8, z.Message("must contain at least 8 characters")).
				Max(72, z.Message("must have at most 72 characters")).
				Required()

	userCreateSchema = z.Struct(z.Schema{
		"email":    z.String().Trim().Max(160).Email(z.Message("must be valid")).Required(),
		"password": passwordFieldSchema,
		"salary":   z.Float().GT(0, z.Message("must be greater than 0")).Required(),
	})

	userLoginSchema = z.Struct(z.Schema{
		"email":    z.String().Trim().Required(),
		"password": z.String().Trim().Required(),
	})

	userResetPasswordSchema = z.Struct(z.Schema{
		"token":    z.String().Trim().Required(),
		"password": passwordFieldSchema,
	})
)

func NewUserHandler(baseHandler *BaseHandler, userService service.UserService) *UserHandler {
	return &UserHandler{
		BaseHandler: baseHandler,
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.With(h.rateLimiter(5, time.Hour)).Post("/users", h.CreateUser)
	r.With(h.rateLimiter(10, 5*time.Minute)).Post("/sessions", h.UserLogin)
	r.With(h.rateLimiter(5, time.Hour)).Post("/password/forgot", h.ForgotPassword)
	r.With(h.rateLimiter(5, time.Hour)).Post("/password/reset", h.ResetPassword)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Salary   float64 `json:"salary"`
	}

	if errs := util.ParseZodSchema(userCreateSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	if _, err := h.userService.GetByEmail(r.Context(), params.Email); err == nil {
		h.sendError(w, http.StatusConflict, "email already in use")
		return
	} else if !errors.Is(err, errs.ErrNotFound{}) {
		h.HandleError(w, err)
		return
	}

	dto := service.CreateUserDTO{
		Email:           params.Email,
		Password:        params.Password,
		CreateSalaryDTO: service.CreateSalaryDTO{Amount: params.Salary},
	}

	user, salary, err := h.userService.Create(r.Context(), dto)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, util.M{
		"user":   user.ToDTO(),
		"salary": salary.ToDTO(),
		"token":  auth.GenerateJWTToken(user.ID, tokenExpiresIn),
	})
}

func (h *UserHandler) UserLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if errs := util.ParseZodSchema(userLoginSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	user, err := h.userService.GetByEmailAndPassword(r.Context(), params.Email, params.Password)
	if errors.Is(err, errs.ErrInvalidCredentials) {
		h.sendError(w, http.StatusUnauthorized, "invalid email or password")
		return
	} else if err != nil {
		panic(err)
	}

	h.sendJSON(w, http.StatusCreated, util.M{"token": auth.GenerateJWTToken(user.ID, tokenExpiresIn)})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		h.HandleError(w, err)
		return
	}

	user, err := h.userService.GetByEmail(r.Context(), params.Email)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	err = h.userService.SendForgotPasswordEmail(r.Context(), *user)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if errs := util.ParseZodSchema(userResetPasswordSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	dto := service.ResetPasswordDTO{
		Token:    params.Token,
		Password: params.Password,
	}

	err := h.userService.ResetPassword(r.Context(), dto)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound{}) {
			h.HandleError(w, errs.NewNotFound("user"))
		} else {
			h.HandleError(w, err)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
