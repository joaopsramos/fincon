package api

import (
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
	*Handler
	userService service.UserService
}

func NewUserHandler(userService service.UserService, handler *Handler) *UserHandler {
	return &UserHandler{
		Handler:     handler,
		userService: userService,
	}
}

var userCreateSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Max(160).Email(z.Message("must be valid")).Required(),
	"password": z.String().Trim().Min(8, z.Message("must contain at least 8 characters")).Max(72, z.Message("must have at most 72 characters")).Required(),
	"salary":   z.Float().GT(0, z.Message("must be greater than 0")).Required(),
})

var userLoginSchema = z.Struct(z.Schema{
	"email":    z.String().Trim().Required(),
	"password": z.String().Trim().Required(),
})

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.With(h.rateLimiter(5, time.Hour)).Post("/users", h.CreateUser)
	r.With(h.rateLimiter(10, 5*time.Minute)).Post("/sessions", h.UserLogin)
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
