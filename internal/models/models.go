package models

type Stat struct {
	ID   uint   `json:"id,omitempty" gorm:"primarykey"`
	Name string `json:"name"         gorm:"unique"`

	SeedPlan float64 `json:"seed_plan"`
	SeedFact float64 `json:"seed_fact"`
	SeedDif  float64 `json:"seed_dif,omitempty"`

	PumpkinPlan float64 `json:"pumpkin_plan"`
	PumpkinFact float64 `json:"pumpkin_fact"`
	PumpkinDif  float64 `json:"pumpkin_dif,omitempty"`

	PeanutPlan float64 `json:"peanut_plan"`
	PeanutFact float64 `json:"peanut_fact"`
	PeanutDif  float64 `json:"peanut_dif,omitempty"`

	AKB1    int `json:"akb1,omitempty"`
	AKB2    int `json:"akb2,omitempty"`
	NewTT   int `json:"newtt,omitempty"`
	Mix     int `json:"mix,omitempty"`
	NpOne   int `json:"npone,omitempty"`
	SetShel int `json:"set_shelving,omitempty"`
	DMP     int `json:"dmp,omitempty"`
	TopFive int `json:"top_five,omitempty"`
	News    int `json:"news,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}

type Password struct {
	Password string `json:"password"`
}
