package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type user struct {
	Id       uint32
	Name     string `json:name`
	Password string `json:password`
}

func jsonError(errar string) gin.H {
	return gin.H{"error": errar}
}

// Attempt to retrieve user information from database
func getUser(db *pgx.Conn, name string) (user, error) {
	var aUser user
	err := db.QueryRow(context.Background(), "select id, name, password from crochet_user where name = $1", name).Scan(&aUser.Id, &aUser.Name, &aUser.Password)

	if err != nil {
		return aUser, errors.New("User does not exist")
	}

	return aUser, nil
}

func register(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newUser user
		if err := ctx.BindJSON(&newUser); err != nil {
			ctx.JSON(http.StatusBadRequest, jsonError("Request is not formatted properly."))
			return
		}

		//Attempt to retrieve user from database. If it works, then the user is already registered.
		if _, err := getUser(db, newUser.Name); err == nil {
			ctx.JSON(http.StatusConflict, jsonError("A user with this name already exists"))
			return
		}

		//Everything good; insert the user
		_, err := db.Query(context.Background(), "insert into crochet_user (name, password) values ($1, $2)", newUser.Name, newUser.Password)
		if err != nil {
			fmt.Printf("%s", err.Error())
			ctx.JSON(http.StatusInternalServerError, jsonError("An error occurred trying to insert the user"))
			return
		}

		ctx.Status(200)
	}
}

func main() {
	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:@localhost:5432/mydb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s", err)
		return
	}
	defer conn.Close(context.Background())

	router := gin.Default()
	router.POST("/register", register(conn))

	router.Run("localhost:8080")
}
