package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

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

type AuthRequest struct {
	ResponseType string `json:"response_type" query:"response_type"`
	ClientID     string `json:"client_id" query:"client_id"`
	RedirectURI  string `json:"redirect_uri" query:"redirect_uri"`
	Scope        string
	State        string
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

	api := fiber.New(fiber.Config{
		AppName: "authorization service",
	})

	api.Use(logger.New())
	api.Use(recover.New())

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello from server!")
	})

	api.Get("/auth", func(c *fiber.Ctx) error {
		authRequest := new(AuthRequest)
		if err := c.QueryParser(authRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request!"})
		}
		return c.SendString("auth!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	api.Listen(fmt.Sprintf(":%s", port))

}
