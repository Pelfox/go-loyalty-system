package repositories

import "errors"

var (
	// ErrUserAlreadyExists указывает на то, что пользователь с таким логином уже
	// зарегистрирован в системе.
	ErrUserAlreadyExists = errors.New("this login already exists")
	// ErrUserNotFound указывает на то, что данный пользователь с таким логином
	// не был найден.
	ErrUserNotFound = errors.New("user with this login does not exist")
	// ErrOrderAlreadyExists указывает на то, что заказ с данным номером уже
	// существует, и не принадлежит данному пользователю.
	ErrOrderAlreadyExists = errors.New("this order already exists")
	// ErrInsufficientFunds используется, если на счету пользователя
	// недостаточно средств.
	ErrInsufficientFunds = errors.New("insufficient funds")
)
