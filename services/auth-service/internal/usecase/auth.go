package usecase

import "context"

type UserRepository interface {
	Ping(context.Context) error
}

type AuthUseCase struct {
	users UserRepository
}

func NewAuthUseCase(users UserRepository) *AuthUseCase {
	return &AuthUseCase{users: users}
}

func (u *AuthUseCase) Health(ctx context.Context) error {
	if u.users == nil {
		return nil
	}
	return u.users.Ping(ctx)
}
