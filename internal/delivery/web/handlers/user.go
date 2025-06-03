package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"ppo/internal/delivery/web/session"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

// UserHandler отвечает за CRUD и аутентификацию пользователей.
type UserHandler struct {
	service services.UserService
	logger  *zap.Logger
}

// NewUserHandler создаёт контроллер для пользователей.
func NewUserHandler(svc services.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		service: svc,
		logger:  log,
	}
}

// RegisterRoutes регистрирует маршруты пользователя.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	// Регистрация и создание
	r.Get("/users/new", h.NewForm)
	r.Post("/users", h.Create)

	// Просмотр / редактирование / удаление
	r.Get("/users/{id}", h.Show)
	r.Get("/users/{id}/edit", h.EditForm)
	r.Post("/users/{id}", h.Update)
	r.Post("/users/{id}/delete", h.Delete)

	// Аутентификация
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
}

// Show показывает детали пользователя по ID.
// GET /users/{id}
func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetUserByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	userDTO := dto.ToUserDTO(user)

	// Парсим только layout.tmpl + user_show.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_show.tmpl",
	))

	data := struct {
		Title string
		User  *dto.UserDTO
	}{
		Title: fmt.Sprintf("Пользователь #%d", user.ID),
		User:  userDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm отображает форму регистрации (создания нового пользователя).
// GET /users/new
func (h *UserHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.CreateUserForm{}

	// Парсим только layout.tmpl + user_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateUserForm
	}{
		Title:      "Регистрация нового пользователя",
		FormAction: "/users",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Create обрабатывает POST /users — создаёт нового пользователя.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	country := r.FormValue("country")
	role := r.FormValue("role")

	cmd := services.RegisterUserCmd{
		Username: username,
		Email:    email,
		Password: password,
		Country:  country,
		Role:     role,
	}

	id, err := h.service.Register(r.Context(), cmd)
	if err != nil {
		h.logger.Error("RegisterUser failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sess, _ := session.Get(r)
	sess.Values["uid"] = id
	sess.Values["role"] = "user" // новый пользователь
	session.Save(r, w, sess)
	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)

	// Перенаправляем на просмотр пользователя
	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)
}

// EditForm отображает форму редактирования пользователя.
// GET /users/{id}/edit
func (h *UserHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetUserByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
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

	// Парсим только layout.tmpl + user_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_form.tmpl",
	))

	data := struct {
		Title      string
		IsNew      bool
		FormAction string
		Form       *dto.UpdateUserForm
	}{
		Title:      fmt.Sprintf("Редактирование пользователя #%d", id),
		IsNew:      false,
		FormAction: fmt.Sprintf("/users/%d", id),
		Form:       form,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Update обрабатывает POST /users/{id} — сохраняет изменения пользователя.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password") // пустой означает «не менять»
	country := r.FormValue("country")
	role := r.FormValue("role")
	isBlocked := r.FormValue("is_blocked") == "on"

	cmd := services.UpdateUserCmd{
		ID:        id,
		Username:  username,
		Email:     email,
		Password:  password,
		Country:   country,
		IsBlocked: isBlocked,
		Role:      role,
	}

	if err := h.service.UpdateUser(r.Context(), cmd); err != nil {
		h.logger.Error("UpdateUser failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)
}

// Delete обрабатывает POST /users/{id}/delete — удаляет пользователя.
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

// LoginForm отображает форму «Вход» (аутентификация).
// GET /login
func (h *UserHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.AuthenticateForm{}

	// Парсим только layout.tmpl + login_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"login_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.AuthenticateForm
	}{
		Title:      "Вход пользователя",
		FormAction: "/login",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Login обрабатывает POST /login — проверяет email+password.
// POST /login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	cmd := services.AuthenticateUserCmd{
		Email:    email,
		Password: password,
	}
	_, err := h.service.Authenticate(r.Context(), cmd)
	if err != nil {
		h.logger.Warn("Authenticate failed", zap.Error(err))
		// Обычно здесь показывают «неверный логин/пароль». Для простоты – 401:
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetByEmail(r.Context(), email) // или тот метод, который возвращает User с полями Role и IsBlocked
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sess, _ := session.Get(r)
	sess.Values["uid"] = user.ID
	sess.Values["role"] = user.Role
	session.Save(r, w, sess)

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)
	sess.Options.MaxAge = -1 // удалить cookie
	session.Save(r, w, sess)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
