package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("пользователь не найден")
	ErrUserAlreadyExists     = errors.New("пользователь с таким номером уже зарегистрирован")
	ErrUsernameAlreadyExists = errors.New("это имя пользователя уже занято")
	ErrInvalidCredentials    = errors.New("неверный номер телефона или пароль")
	ErrInvalidInput          = errors.New("некорректные входные данные")
	ErrInternal              = errors.New("внутренняя ошибка сервера")
)
