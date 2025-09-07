package auth

import (
	"net/smtp"
	"time"

	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/subosito/gotenv"
	"golang.org/x/crypto/bcrypt"
)

func ForgotPassword(c *fiber.Ctx) error {
	gotenv.Load()
	u := new(models.PasswordReset)

	if err := c.BodyParser(&u); err != nil {
		return err
	}

	token := utils.GenerateRandomString(12)

	pr := &models.PasswordReset{
		Email: u.Email,
		Token: token,
	}

	// search for the email in the database, if the user exist
	um := &models.User{}

	database.DB.Where("email = ?", u.Email).First(um)
	if um.UUID == "" {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "invalid email address 😰",
		})
	}

	// token expiration time is 3hr
	pr.ExpirationTime = time.Now().Add(time.Hour * time.Duration(3))
	pr.CreatedAt = time.Now()

	database.DB.Create(pr)

	from := utils.Env("EMAIL_FROM")

	to := []string{
		u.Email,
	}

	auth := smtp.PlainAuth("", utils.Env("EMAIL_USERNAME"), utils.Env("EMAIL_PASSWORD"), utils.Env("EMAIL_HOST"))

	url := utils.Env("RESET_URL") + token

	msg := []byte("Click <a href=\"" + url + "\">here</a> to reset your password!")

	err := smtp.SendMail(utils.Env("EMAIL_HOST")+":"+utils.Env("EMAIL_PORT"), auth, from, to, msg)
	if err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "email was not sent 😰",
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
	})

}

func ResetPassword(c *fiber.Ctx) error {

	rp := &models.PasswordReset{}

	if err := database.DB.Where("token = ?", c.Params("token")).Last(rp); err.Error != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "invalid token",
		})
	}

	if rp.UUID == "" {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "invalid token",
		})
	}

	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	if now.After(rp.ExpirationTime) {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "token has expired",
		})
	}

	r := new(models.Reset)

	if r.Password != r.PasswordConfirm {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "password does not match",
		})
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(r.Password), 14)
	database.DB.Model(&models.User{}).Where("email = ?", rp.Email).Update("password", password)

	return c.JSON(fiber.Map{
		"message": "success",
	})

}
