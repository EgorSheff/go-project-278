package dto

type Link struct {
	Id          int32  `json:"id"`
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
	ShortURL    string `json:"short_url"`
}
