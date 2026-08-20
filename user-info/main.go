package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	common "crochet.com/common"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var userQueries = map[string]string{
	"from_name":      "select id from crochet_user where name = $1",
	"add":            "insert into crochet_user (name, password) values ($1, $2) returning id",
	"get_patterns":   "select id from crochet_pattern where creator_id = $1",
	"get_progresses": "select id from crochet_progress where user_id = $1",
}

func jsonError(errar string) gin.H {
	return gin.H{"error": errar}
}

// Attempt to retrieve user information from database
func getUserFromName(db *pgx.Conn, name string) (uint64, error) {
	var userID uint64
	err := db.QueryRow(context.Background(), userQueries["from_name"], name).Scan(&userID)

	if err != nil {
		return 0, errors.New("User does not exist")
	}

	return userID, nil
}

// Register user
func register(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newUser common.User
		if err := ctx.BindJSON(&newUser); err != nil {
			ctx.JSON(http.StatusBadRequest, jsonError("Request is not formatted properly."))
			return
		}

		//Attempt to retrieve user from database. If it works, then the user is already registered.
		if _, err := getUserFromName(db, newUser.Name); err == nil {
			ctx.JSON(http.StatusConflict, jsonError("A user with this name already exists"))
			return
		}

		//Everything good; insert the user
		row, err := db.Query(context.Background(), userQueries["add"], newUser.Name, newUser.Password)
		if err != nil {
			log.Fatalln(err.Error())
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
			log.Fatalln(err.Error())
			ctx.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
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
		rows, err := db.Query(context.Background(), userQueries["get_patterns"], uid)
		if err != nil {
			log.Fatalln(err.Error())
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
			log.Fatalln(err.Error())
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
		rows, err := db.Query(context.Background(), userQueries["get_progresses"], uid)
		if err != nil {
			log.Fatalln(err.Error())
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
			log.Fatalln(err.Error())
			c.JSON(http.StatusInternalServerError, jsonError("An internal server error occurred"))
			return
		}

		c.JSON(http.StatusOK, gin.H{"progress": progress})
	}
}

func main() {
	//Parse Postgres info
	postgres, err := common.NewDatabaseFromEnv[common.PostgresInfo]()
	if err != nil {
		log.Fatal(err.Error())
	}

	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), postgres.GetUrl())
	if err != nil {
		log.Fatalf("Could not connect to postgres :( %s\n", err)
		return
	}
	defer conn.Close(context.Background())

	router := gin.Default()
	router.POST("/register", register(conn))
	router.GET("/patterns/:id", created_patterns(conn))
	router.GET("/progress/:id", created_progress(conn))

	router.Run("0.0.0.0:8080")
}
