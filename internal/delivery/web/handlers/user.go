package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"ppo/internal/repositories"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/middleware"
	"ppo/internal/delivery/web/session"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

type UserHandler struct {
	service services.UserService
	logger  *zap.Logger
}

func NewUserHandler(svc services.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		service: svc,
		logger:  log,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Get("/users/new", h.NewForm)
	r.Post("/users", h.Create)

	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	r.With(middleware.RequireRole("admin")).Get("/users", h.List)

	r.With(middleware.RequireRole("user", "admin")).Get("/users/{id}", h.Show)
	r.With(middleware.RequireRole("user", "admin")).Get("/users/{id}/edit", h.EditForm)
	r.With(middleware.RequireRole("user", "admin")).Post("/users/{id}", h.Update)
	r.With(middleware.RequireRole("user", "admin")).Get("/profile", h.ProfileShow)

	r.With(middleware.RequireRole("admin")).Post("/users/{id}/delete", h.Delete)
}

func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if currentRole != "admin" && currentUID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	userDTO := dto.ToUserDTO(user)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_show.tmpl",
	))

	data := struct {
		Title  string
		Role   string
		UserID uint64
		User   *dto.UserDTO
	}{
		Title:  fmt.Sprintf("Пользователь #%d", user.ID),
		Role:   currentRole,
		UserID: currentUID,
		User:   userDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *UserHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_form.tmpl",
	))

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		IsNew      bool
		FormAction string
		Form       *dto.CreateUserForm
	}{
		Title:      "Регистрация нового пользователя",
		Role:       "guest",
		UserID:     0,
		IsNew:      true,
		FormAction: "/users",
		Form:       &dto.CreateUserForm{},
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	cmd := services.RegisterUserCmd{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Country:  r.FormValue("country"),
		Role:     "user",
	}

	newID, err := h.service.Register(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		h.logger.Error("RegisterUser failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sess, _ := session.Get(r)
	sess.Values["uid"] = newID
	sess.Values["role"] = "user"
	session.Save(r, w, sess)

	http.Redirect(w, r, fmt.Sprintf("/users/%d", newID), http.StatusSeeOther)
}

func (h *UserHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if currentRole != "admin" && currentUID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	form := &dto.UpdateUserForm{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Country:   user.Country,
		IsBlocked: user.IsBlocked,
		Role:      user.Role,
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_form.tmpl",
	))

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		IsNew      bool
		FormAction string
		Form       *dto.UpdateUserForm
	}{
		Title:      fmt.Sprintf("Редактирование пользователя #%d", id),
		Role:       currentRole,
		UserID:     currentUID,
		IsNew:      false,
		FormAction: fmt.Sprintf("/users/%d", id),
		Form:       form,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if currentRole != "admin" && currentUID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	cmd := services.UpdateUserCmd{
		ID:       id,
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"), // если пусто, сервис игнорирует
		Country:  r.FormValue("country"),
	}

	if currentRole == "admin" {
		cmd.Role = r.FormValue("role")
		cmd.IsBlocked = (r.FormValue("is_blocked") == "on")
	}

	if err := h.service.UpdateUser(r.Context(), cmd); err != nil {
		h.logger.Error("UpdateUser failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		h.logger.Error("DeleteUser failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *UserHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"login_form.tmpl",
	))

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		FormAction string
		Form       *dto.AuthenticateForm
	}{
		Title:      "Вход пользователя",
		Role:       "guest",
		UserID:     0,
		FormAction: "/login",
		Form:       &dto.AuthenticateForm{},
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	_ = currentUID

	users, err := h.service.ListAllUsers(r.Context())
	if err != nil {
		h.logger.Error("ListAllUsers failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dtos := dto.ToUserDTOs(users)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_list.tmpl",
	))

	data := struct {
		Title  string
		Role   string
		UserID uint64
		Users  []*dto.UserDTO
	}{
		Title:  "Список пользователей",
		Role:   currentRole,
		UserID: currentUID,
		Users:  dtos,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cmd := services.AuthenticateUserCmd{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	user, err := h.service.Authenticate(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, services.ErrUserBlocked) {
			http.Error(w, "Your account is blocked", http.StatusForbidden)
			return
		}
		h.logger.Error("Authenticate failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sess, _ := session.Get(r)
	sess.Values["uid"] = user.ID
	sess.Values["role"] = user.Role
	session.Save(r, w, sess)

	http.Redirect(w, r, "/users/"+strconv.FormatUint(user.ID, 10), http.StatusSeeOther)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)
	sess.Options.MaxAge = -1
	session.Save(r, w, sess)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *UserHandler) ProfileShow(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	user, err := h.service.GetUserByID(r.Context(), currentUID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	userDTO := dto.ToUserDTO(user)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"profile_show.tmpl",
	))

	data := struct {
		Title  string
		Role   string
		UserID uint64
		User   *dto.UserDTO
	}{
		Title:  "Личный кабинет",
		Role:   currentRole,
		UserID: currentUID,
		User:   userDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}
