package domain

type Endpoint struct {
	Name  string `gorm:"primary_key;type:text;not null;unique" json:"name"`
	Group string `gorm:"type:text;" json:"group"`
}
