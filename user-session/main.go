package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type user struct {
	Id       uint32
	Name     string `json:name`
	Password string `json:password`
}

type session struct {
	SessionID  string
	UserID     uint32
	LastUpdate time.Time
}

// User ID -> Session
var activeUsers map[uint32]string

// Session ID -> Session
var sessions map[string]session

// Get a fresh, new session key
var sessionCount uint64

func getSessionKey() string {
	return strconv.FormatUint(sessionCount, 10)
}

// Authenticate the user
func authUser(db *pgx.Conn, username string, password string) (user, error) {
	//Attempt to find user with given password
	user := user{}

	err := db.QueryRow(context.Background(), "select id, name, password from crochet_user where name = $1 and password = $2", username, password).Scan(&user.Id, &user.Name, &user.Password)
	if err != nil {
		return user, fmt.Errorf("user %s does not exist or wrong password was used", username)
	}

	return user, nil
}

// Add a user to the session
func addUser(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get user login
		var user user

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "request does not contain username and password"})
			return
		}

		//Authenticate user
		user, err := authUser(db, user.Name, user.Password)
		if err != nil {
			c.JSON(404, map[string]any{"error": err.Error()})
		}

		//Check if the user already has a sesson. If so, stop
		if _, ok := activeUsers[user.Id]; ok == true {
			c.JSON(http.StatusTeapot, map[string]any{"error": "user already has an active session"})
			return
		}

		//Insert them into the session
		sessionKey := getSessionKey()
		sessions[sessionKey] = session{
			UserID:     user.Id,
			SessionID:  sessionKey,
			LastUpdate: time.Now(),
		}
		activeUsers[user.Id] = sessionKey

		//Return status
		c.String(200, sessionKey)
	}
}

func main() {
	//Make sessions/activeUsers map
	sessions = make(map[string]session)
	activeUsers = make(map[uint32]string)
	sessionCount = 0

	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:@localhost:5432/mydb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s", err)
		return
	}
	defer conn.Close(context.Background())

	router := gin.Default()
	router.POST("/login", addUser(conn))

	router.Run("localhost:8080")
}
