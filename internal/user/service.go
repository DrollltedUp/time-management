package user

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Register(email, password, fullname string) (User, error) {

	return User{}, nil
}
