package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	common "crochet.com/common"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var patternQueries = map[string]string{
	"exist":  "select id from crochet_pattern where name=$1 and creator_id=$2",
	"insert": "insert into crochet_pattern (name, steps, creator_id) values ($1, $2, $3) returning id",
	"get":    "select * from crochet_pattern where id=$1",
}

// Get pattern from name and creator id
func getPattern(db *pgx.Conn, name string, creatorId uint64) (uint64, error) {
	var id uint64
	err := db.QueryRow(context.Background(), patternQueries["exist"], name, creatorId).Scan(&id)
	return id, err
}

// Create a new pattern
func createPattern(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newPattern common.Pattern
		if err := ctx.BindJSON(&newPattern); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			ctx.JSON(http.StatusBadRequest, common.JsonError("Request is not formatted properly."))
			return
		}

		fmt.Println(newPattern)

		//Attempt to retrieve pattern from database. If it works, then the pattern is already created.
		if _, err := getPattern(db, newPattern.Name, newPattern.Id); err == nil {
			ctx.JSON(http.StatusConflict, common.JsonError("A pattern with this name already exists"))
			return
		}

		//Everything good; create the pattern
		row, err := db.Query(context.Background(), patternQueries["insert"], newPattern.Name, newPattern.Steps, newPattern.Creator)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			ctx.JSON(http.StatusInternalServerError, common.JsonError("An error occurred trying to create the pattern"))
			return
		}

		//Attempt to retrieve id
		id, err := pgx.CollectExactlyOneRow(row, func(row pgx.CollectableRow) (int32, error) {
			var i int32
			err := row.Scan(&i)
			return i, err
		})

		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			ctx.JSON(http.StatusInternalServerError, common.JsonError("An internal server error occurred"))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"id": id})
	}
}

// Get pattern details for a given pattern ID
func getPatternDetails(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		patternId, exists := c.Params.Get("id")
		if !exists {
			c.JSON(http.StatusBadRequest, common.JsonError("Request was not formatted properly"))
			return
		}

		//Retrieve pattern from its id
		var pattern common.Pattern
		err := db.QueryRow(context.Background(), patternQueries["get"], patternId).Scan(&pattern.Id, &pattern.Name, &pattern.Steps, &pattern.Creator)

		if err != nil {
			c.JSON(http.StatusNotFound, common.JsonError("A pattern with that ID could not be found"))
			return
		}

		//Return pattern

		c.JSON(http.StatusOK, pattern)
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
	router.POST("/create", createPattern(conn))
	router.GET("/:id", getPatternDetails(conn))

	router.Run("0.0.0.0:8080")
}
