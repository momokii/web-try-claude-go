package middlewares

import (
	"log"
	"os"
	"time"

	"scrapper-test/database"
	"scrapper-test/models"
	sessionRepo "scrapper-test/repository/session"
	"scrapper-test/repository/user"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var (
	Store   *session.Store
	SSO_URL = os.Getenv("SSO_URL")
)

func InitSession() {
	Store = session.New(session.Config{
		Expiration:     7 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		// change cookie name to session_id_tryllm for not overlapping with sso session
		CookieName: "session_id_tryllm",
		KeyLookup:  "cookie:session_id_tryllm",
	})

	log.Println("Session store initialized")
}

func CreateSession(c *fiber.Ctx, key string, value interface{}) error {
	sess, err := Store.Get(c)
	if err != nil {
		return err
	}
	defer sess.Save()

	sess.Set(key, value)

	return nil
}

func DeleteSession(c *fiber.Ctx) error {
	sess, err := Store.Get(c)
	if err != nil {
		return err
	}
	defer sess.Save()

	sess.Destroy()

	return nil
}

func CheckSession(c *fiber.Ctx, key string) (interface{}, error) {
	sess, err := Store.Get(c)
	if err != nil {
		return nil, err
	}

	return sess.Get(key), nil
}

func IsNotAuth(c *fiber.Ctx) error {
	userid, err := CheckSession(c, "id")
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	session_id, err := CheckSession(c, "session_id")
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	if userid != nil && session_id != nil {
		return c.Redirect("/")
	}

	return c.Next()
}

func IsAuth(c *fiber.Ctx) error {
	userid, err := CheckSession(c, "id")
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	session_id, err := CheckSession(c, "session_id")
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	// if session data not found, redirect to login
	if userid == nil || session_id == nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}
	defer func() {
		database.CommitOrRollback(tx, c, err)
	}()

	// check if session is valid
	userRepo := user.NewUserRepo()
	session_repo := sessionRepo.NewSessionRepo()

	// first check if session is valid on database
	sessData, err := session_repo.FindSession(tx, session_id.(string), userid.(int))
	// if session not found or error happen, redirect to login and delete the session local data
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	// if session is deleted/ not found
	if sessData.Id == 0 && sessData.UserId == 0 && sessData.SessionId == "" {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	userData, err := userRepo.FindByID(tx, userid.(int))
	if err != nil {
		DeleteSession(c)
		return c.Redirect(SSO_URL)
	}

	userSession := models.UserSession{
		Id:       userData.Id,
		Username: userData.Username,
	}

	// store information for next data
	c.Locals("user", userSession)

	return c.Next()
}
