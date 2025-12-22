package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/lucsky/cuid"

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

type ConfirmAuthRequest struct {
	Authorize bool `json:"authorize" query:"authorize"`
	State     string
	ClientID  string `json:"client_id" query:"client_id"`
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
		ID:          "19",
		Name:        "fibers",
		Website:     "https://gofiber.io",
		Logo:        "https://avatars.githubusercontent.com/u/40920169?s=200&v=4",
		RedirectURI: "https://localhost:8080/callback",
	})

	views := html.New("./views", ".html")

	api := fiber.New(fiber.Config{
		AppName: "Authorization Service",
		Views:   views,
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

		if authRequest.ResponseType != "code" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		if authRequest.ClientID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		if !strings.Contains(authRequest.RedirectURI, "https") {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		if authRequest.Scope == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		if authRequest.State == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		//Check for client
		client := new(Client)
		if err := db.Where("name = ?", authRequest.ClientID).First(&client).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_client"})
		}

		//Generate temp code
		code, err := cuid.NewCrypto(rand.Reader)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "server_error"})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "temp_auth_request_code",
			Value:    code,
			Secure:   true,
			Expires:  time.Now().Add(1 * time.Minute),
			HTTPOnly: true,
		})

		return c.Render("authorize_client", fiber.Map{
			"Logo":    client.Logo,
			"Name":    client.Name,
			"Website": client.Website,
			"State":   authRequest.State,
			"Scopes":  strings.Split(authRequest.Scope, " "),
		})
	})

	api.Get("/confirm_auth", func(c *fiber.Ctx) error {
		tempCode := c.Cookies("temp_auth_request_code")
		if tempCode == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		ConfirmAuthRequest := new(ConfirmAuthRequest)
		if err := c.QueryParser(ConfirmAuthRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
		}

		//Check for client
		client := new(Client)
		if err := db.Where("name = ?", ConfirmAuthRequest.ClientID).First(&client).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid_client"})
		}

		if !ConfirmAuthRequest.Authorize {
			return c.Redirect(client.RedirectURI + "?error=access_denied" + "&state=" + ConfirmAuthRequest.State)
		}

		return c.Redirect(client.RedirectURI + "?code=" + tempCode + "&state=" + ConfirmAuthRequest.State)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	api.Listen(fmt.Sprintf(":%s", port))

}
