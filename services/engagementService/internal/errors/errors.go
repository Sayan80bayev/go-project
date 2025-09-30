package errors

import "errors"

// Common errors
var (
	ErrInvalidArgument       = errors.New("invalid argument")
	ErrNotFound              = errors.New("not found")
	ErrDuplicateSubscription = errors.New("subscription already exists")
	ErrAlreadyFollowing      = errors.New("already following")
	ErrNotFollowing          = errors.New("not following")
	ErrDuplicateLike         = errors.New("like already exists for user and post")
)