package dto

import "time"

type Visit struct {
	ID        int32     `json:"id"`
	LinkID    int32     `json:"link_id"`
	CreatedAt time.Time `json:"created_at"`
	Ip        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Status    int16     `json:"status"`
}
