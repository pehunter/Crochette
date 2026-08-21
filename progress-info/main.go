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

var progressQueries = map[string]string{
	"exist":  "select id from crochet_progress where pattern_id=$1 and user_id=$2",
	"insert": "insert into crochet_progress (progress, pattern_id, user_id) values (0, $1, $2) returning id",
	"get":    "select * from crochet_progress where id=$1",
	"update": "update crochet_progress set progress=$1 where id=$2",
}

// Get progress for a pattern under a user
func getProgress(db *pgx.Conn, patternId uint64, userId uint64) (uint64, error) {
	var id uint64
	err := db.QueryRow(context.Background(), progressQueries["exist"], patternId, userId).Scan(&id)
	return id, err
}

// Create a new progress
func createProgress(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newProgress common.Progress
		if err := ctx.BindJSON(&newProgress); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			ctx.JSON(http.StatusBadRequest, common.JsonError("Request is not formatted properly."))
			return
		}

		fmt.Println(newProgress.Pattern)
		fmt.Println(newProgress.User)
		
		//Attempt to retrieve progress from database. If it works, then the progress is already created.
		if _, err := getProgress(db, newProgress.Pattern, newProgress.User); err == nil {
			ctx.JSON(http.StatusConflict, common.JsonError("A progress with this name already exists"))
			return
		}

		fmt.Println(newProgress.Pattern)
		fmt.Println(newProgress.User)

		//Everything good; create the progress
		row, err := db.Query(context.Background(), progressQueries["insert"], newProgress.Pattern, newProgress.User)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			ctx.JSON(http.StatusInternalServerError, common.JsonError("An error occurred trying to create a progress"))
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

// Get progress details for a given progress ID
func getProgressDetails(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		progressId, exists := c.Params.Get("id")
		if !exists {
			c.JSON(http.StatusBadRequest, common.JsonError("Request was not formatted properly"))
			return
		}

		//Retrieve progress from its id
		var progress common.Progress
		err := db.QueryRow(context.Background(), progressQueries["get"], progressId).Scan(&progress.Id, &progress.Progression, &progress.Pattern, &progress.User)

		if err != nil {
			c.JSON(http.StatusNotFound, common.JsonError("A progress with that ID could not be found"))
		}

		//Return progress
		c.JSON(http.StatusOK, progress)
	}
}

// Update progress progression
func updateProgress(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := struct {
			id          uint64
			progression uint64
		}{
			id:          0,
			progression: 0,
		}

		if err := c.ShouldBind(&request); err != nil {
			c.JSON(http.StatusBadRequest, common.JsonError("Request was not formatted properly"))
			return
		}

		//Attempt to update progress
		err := db.QueryRow(context.Background(), progressQueries["update"], request.id, request.progression).Scan()

		if err != nil {
			c.JSON(http.StatusNotFound, common.JsonError("A progress with that ID could not be found"))
		}

		//Return progress
		c.Status(http.StatusOK)
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
	router.POST("/create", createProgress(conn))
	router.POST("/update", updateProgress(conn))
	router.GET("/:id", getProgressDetails(conn))

	router.Run("0.0.0.0:8080")
}
