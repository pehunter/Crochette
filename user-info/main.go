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
		row, err := db.Query(context.Background(), "insert into crochet_user (name, password) values ($1, $2) returning id", newUser.Name, newUser.Password)
		if err != nil {
			fmt.Printf("%s", err.Error())
			ctx.JSON(http.StatusInternalServerError, jsonError("An error occurred trying to insert the user"))
			return
		}

		//Attempt to retrieve id
		id, err := pgx.CollectExactlyOneRow(row, func(row pgx.CollectableRow) (int32, error) {
			var i int32
			err := row.Scan(&i)
			return i, err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			ctx.JSON(http.StatusInternalServerError, jsonError("An internal serfver error occurred"))
			return
		}

		ctx.JSON(200, gin.H{"id": id})
	}
}

// Get all patterns IDs created by a user
func created_patterns(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		//Check ID param
		uid, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, jsonError("No User ID was given"))
			return
		}

		//Attempt to fetch rows
		rows, err := db.Query(context.Background(), "select id from crochet_pattern where creator_id = $1", uid)
		if err != nil {
			fmt.Printf("%s", err.Error())
			c.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
			return
		}

		//Collect rows
		patterns, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (int32, error) {
			var id int32
			err := row.Scan(&id)
			return id, err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			c.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
			return
		}

		c.JSON(http.StatusOK, gin.H{"patterns": patterns})
	}
}

// Get all patterns IDs created by a user
func created_progress(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {

		//Check ID param
		uid, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, jsonError("No User ID was given"))
			return
		}

		//Attempt to fetch rows
		rows, err := db.Query(context.Background(), "select id from crochet_progress where user_id = $1", uid)
		if err != nil {
			fmt.Printf("%s", err.Error())
			c.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
			return
		}

		//Collect rows
		progress, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (int32, error) {
			var id int32
			err := row.Scan(&id)
			return id, err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			c.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
			return
		}

		c.JSON(http.StatusOK, gin.H{"progress": progress})
	}
}

func main() {
	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:password@postgresql:5432/mydb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s", err)
		return
	}
	defer conn.Close(context.Background())

	router := gin.Default()
	router.POST("/register", register(conn))
	router.GET("/patterns/:id", created_patterns(conn))
	router.GET("/progress/:id", created_progress(conn))

	router.Run("0.0.0.0:8080")
}
