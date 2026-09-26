package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"stack/internal/domain"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req domain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	resp, err := h.authUsecase.Register(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req domain.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	resp, err := h.authUsecase.Login(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		respondWithError(w, http.StatusConflict, "Пользователь с таким номером уже зарегистрирован")
	case errors.Is(err, domain.ErrUsernameAlreadyExists):
		respondWithError(w, http.StatusConflict, "Этот username уже занят")
	case errors.Is(err, domain.ErrInvalidCredentials):
		respondWithError(w, http.StatusUnauthorized, "Неверный телефон или пароль")
	default:
		respondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
}
