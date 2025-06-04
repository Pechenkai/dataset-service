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
	// Регистрация — для всех (guest тоже)
	r.Get("/users/new", h.NewForm)
	r.Post("/users", h.Create)

	// Логин/Logout
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	r.With(middleware.RequireRole("admin")).Get("/users", h.List)

	// Просмотр/редактирование своего профиля
	r.With(middleware.RequireRole("user", "admin")).Get("/users/{id}", h.Show)
	r.With(middleware.RequireRole("user", "admin")).Get("/users/{id}/edit", h.EditForm)
	r.With(middleware.RequireRole("user", "admin")).Post("/users/{id}", h.Update)

	// Удаление — только админ
	r.With(middleware.RequireRole("admin")).Post("/users/{id}/delete", h.Delete)
}

// Show показывает детали пользователя по ID.
// GET /users/{id}
func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	// 1) Извлечём из контекста: кто залогинен
	currentUID, currentRole := middleware.FromContext(r.Context())

	// 2) Параметр {id}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 3) Если роль не admin и пытаются смотреть чужого юзера — Forbidden
	if currentRole != "admin" && currentUID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 4) Получаем из сервиса
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

// NewForm отображает форму регистрации (создание нового пользователя).
// GET /users/new
func (h *UserHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"user_form.tmpl",
	))

	// Роль гостя — "guest", UID=0
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

// Create обрабатывает POST /users — регистрация.
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
		Role:     "user", // на форме не даём выбирать роль — всегда "user"
	}

	// Возвращает новый ID или ошибку
	newID, err := h.service.Register(r.Context(), cmd)
	if err != nil {
		// Если почта занята, сервис вернёт repositories.ErrEmailAlreadyExists
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		h.logger.Error("RegisterUser failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// После успешной регистрации — сразу логиним пользователя
	sess, _ := session.Get(r)
	sess.Values["uid"] = newID
	sess.Values["role"] = "user"
	session.Save(r, w, sess)

	http.Redirect(w, r, fmt.Sprintf("/users/%d", newID), http.StatusSeeOther)
}

// EditForm отображает форму редактирования пользователя.
// GET /users/{id}/edit
func (h *UserHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Проверка: либо админ, либо редактируем свой профиль
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

// Update обрабатывает POST /users/{id} — сохраняет изменения.
// POST /users/{id}
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

// Delete обрабатывает POST /users/{id}/delete — удаляет пользователя.
// (маршрут обёрнут в RequireRole("admin") => сюда может попасть только админ)
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

// LoginForm отображает форму входа.
// GET /login
func (h *UserHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"login_form.tmpl",
	))

	// Здесь роль=guest, userID=0
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

// List показывает страницу со списком всех пользователей. (GET /users)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	// currentRole тут гарантированно "admin" (middleware.RequireRole)
	_ = currentUID

	users, err := h.service.ListAllUsers(r.Context())
	if err != nil {
		h.logger.Error("ListAllUsers failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Конвертируем в DTO
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

// Login обрабатывает POST /login — проверяет email+password.
// В случае успеха сохраняет role и uid в сессии.
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

// Logout обрабатывает POST /logout — удаляет сессию.
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)
	sess.Options.MaxAge = -1
	session.Save(r, w, sess)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
