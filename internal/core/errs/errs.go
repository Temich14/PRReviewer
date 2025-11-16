package errs

import (
	"errors"
)

var ErrAlreadyExists = errors.New("сущность с такими параметрами уже существует")
var ErrNotFound = errors.New("сущность не найдена")
var ErrNoReviewersAvailable = errors.New("нет доступных ревьюеров")
var ErrUserNotAssigned = errors.New("пользователь не был назначен ревьюером")
var ErrAlreadyMerged = errors.New("cannot reassign on merged PR")
