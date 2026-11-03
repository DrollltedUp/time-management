package shifts

import "time"

type ShiftDTO struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	StoreID  int64     `json:"store_id"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
	Type     DayNight  `json:"type"`
}

func ToDTO(s Shift) ShiftDTO {
	return ShiftDTO{
		ID:       s.ID,
		UserID:   s.UserID,
		StoreID:  s.StoreID,
		StartsAt: s.StartsAt,
		EndsAt:   s.EndsAt,
		Type:     s.Type(),
	}
}

type ShiftWithUSerDTO struct {
	ShiftDTO
	FullName string `json:"full_name"`
}

type ShiftWithStoreDTO struct {
	ShiftDTO
	StoreNumber  string `json:"store_number"`
	StoreAddress string `json:"store_address"`
}
