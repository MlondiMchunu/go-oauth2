package main

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Client struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Website     string
	Logo        string
	RedirectURI string         `json:"redirect_uri"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbUrl := os.Getenv("DATABASE_URI")
	if dbUrl == "" {
		panic("DATABASE_URI not set!")
	}

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database!")
	}

	//migrate the schema
	db.AutoMigrate(&Client{})

	//insert dummy client
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "website", "redirect_uri", "logo"}),
	}).Create(&Client{
		ID:          "15",
		Name:        "fiber",
		Website:     "https://gofiber.io",
		Logo:        "https://avatars.githubusercontent.com/u/40920169?s=200&v=4",
		RedirectURI: "http://localhost:3000/callback",
	})

}
