package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type user struct {
	Name     string `json:name`
	Password string `json:password`
}

type session struct {
	SessionID  string `json:"key"`
	UserID     string `json:"key"`
	LastUpdate string `json:"last_update"`
}

// User ID -> Session
var activeUsers map[string]string

// Session ID -> Session
var sessions map[string]session

// Add a user to the session
func addUser(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get user login
		var user user

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "request does not contain username and password"})
			return
		}

		//Attempt to find user with given password
		var name string
		var password string
		err := db.QueryRow(context.Background(), "select name, password from crochet_user where name = $1 and password = $2", user.Name, user.Password).Scan(&name, &password)
		if err != nil {
			errMsg := fmt.Sprintf("(am i being gaslit?) user %s does not exist or wrong password was used", user.Name)
			c.JSON(http.StatusNotFound, map[string]string{"error": errMsg})
			return
		}

		//Check if the user already has a sesson. If so, stop
		if _, ok := activeUsers[name]; ok == true {
			c.JSON(http.StatusTeapot, map[string]any{"error": "user already has an active session"})
			return
		}

		//Return status
		c.String(200, "It worked?")
	}
}

func main() {
	//Make sessions/activeUsers map
	sessions = make(map[string]session)
	activeUsers = make(map[string]string)

	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:@localhost:5432/postgres")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s", err)
		return
	}
	defer conn.Close(context.Background())

	router := gin.Default()
	router.POST("/login", addUser(conn))

	router.Run("localhost:8080")
}
