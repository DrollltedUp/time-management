package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Register(ctx context.Context, email, password, fullname string) (User, error) {

	normalizeEmail := lowerEmail(email)
	normalFullName := normalizeName(fullname)

	if !isValidEmail(normalizeEmail) || !isValidPassword(password) || !validName(normalFullName) {
		return User{}, ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		Email:         normalizeEmail,
		Password_hash: string(hash),
		Fullname:      normalFullName,
		Role:          RoleEmployee,
		Active:        true,
	}

	created, err := s.store.Create(ctx, user)
	if err != nil {
		return User{}, ErrEmailTaken
	}

	return created, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	normalizadeEmail := lowerEmail(email)

	usermail, err := s.store.GetByMail(ctx, normalizadeEmail)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrInvalidInput
		}
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(usermail.Password_hash), []byte(password))
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	return usermail, nil
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func isValidPassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8 && len(password) <= 72
}

func lowerEmail(email string) string {
	return strings.ToLower(email)
}

func normalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

func validName(n string) bool {
	c := utf8.RuneCountInString(n)
	return c >= 1 && c <= 100
}
