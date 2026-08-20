package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	common "crochet.com/common"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// Get a fresh, new session key
var sessionCount uint64

func getSessionKey() string {
	sessionCount += 1
	return strconv.FormatUint(sessionCount, 10)
}


// Refresh a session's last update
func refreshSession(reds *redis.Client, sessionKey string) error {
	_, err := reds.ZAdd(context.Background(), "last_accessed", redis.Z{Member: sessionKey, Score: float64(time.Now().Unix())}).Result()

	if err != nil {
		return err
	}

	return nil
}

// Cache a session in redis
func cacheSession(reds *redis.Client, sessionKey string, uid uint64) error {
	_, err := reds.HSet(context.Background(), "session_to_user", sessionKey, strconv.FormatUint(uid, 10)).Result()

	if err != nil {
		return err
	}

	_, err = reds.HSet(context.Background(), "user_to_session", strconv.FormatUint(uid, 10), sessionKey).Result()

	if err != nil {
		return err
	}

	return refreshSession(reds, sessionKey)
}

// Add a user to the session
func addUser(db *pgx.Conn, reds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get user login
		var user common.User

		if err := c.BindJSON(&user); err != nil {
			c.String(http.StatusBadRequest, "Request has missing parameters.")
			return
		}

		//Authenticate user
		user, err := common.NewUserFromAuth(db, user.Name, user.Password)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		//Check if the user already has a sesson. If so, stop
		if _, err := reds.HGet(context.Background(), "user_to_session", strconv.FormatUint(user.Id, 10)).Result(); err == nil {
			c.String(http.StatusUnauthorized, "User has already logged in.")
			return
		}

		//Insert them into the session
		sessionKey := getSessionKey()
		err = cacheSession(reds, sessionKey, user.Id)
		if err != nil {
			log.Print("here?\n")
			log.Fatalf("%s\n", err.Error())
		}

		//Return status
		c.String(http.StatusOK, sessionKey)
	}
}

// Get the user associated with the session
func sessionUser(reds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, ok := c.Params.Get("session")
		if !ok {
			c.String(http.StatusBadRequest, "No session key provided.")
			return
		}

		if uid, err := reds.HGet(context.Background(), "session_to_user", sessionKey).Result(); err != nil {
			c.String(http.StatusNotFound, "The session key has expired.")
		} else {
			// fmt.Fprintf(os.Stdout, "%s\n", session.LastUpdate.String())

			//Refresh session end time
			refreshSession(reds, sessionKey)

			//Return user ID
			c.String(http.StatusOK, uid)
		}
	}
}

// How long a session will last
var sessionTimeout time.Duration

// Remove old sessions
func clearOldSessions(reds *redis.Client) {
	//Delete out-of-date entries
	cutoff := time.Now().Add(-sessionTimeout).Unix()
	_, err := reds.ZRemRangeByScore(context.Background(), "last_accessed", strconv.FormatInt(cutoff, 10), strconv.FormatInt(time.Now().Unix(), 10)).Result()

	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
}

func main() {
	//Parse Postgres & Redis info
	postgres, err := common.NewDatabaseFromEnv[common.PostgresInfo]()
	if err != nil {
		log.Fatal(err.Error())
	}

	reds, err := common.NewDatabaseFromEnv[common.RedisInfo]()
	if err != nil {
		log.Fatal(err.Error())
	}

	//Make sessions/activeUsers map
	sessionCount = 0
	sessionTimeout = time.Second * 18

	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), postgres.GetUrl())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s\n", err)
		return
	}
	defer conn.Close(context.Background())

	//Connect to Redis
	rConn := redis.NewClient(&redis.Options{
		Addr:     reds.GetUrl(),
		Password: "", // no password docs
		DB:       0,  // use default DB
		Protocol: 2,
	})

	//Start session garbage collection
	go func() {
		for {
			clearOldSessions(rConn)
			time.Sleep(time.Second * 5)
		}
	}()

	router := gin.Default()
	router.POST("/login", addUser(conn, rConn))
	router.GET("/:session", sessionUser(rConn))

	router.Run("0.0.0.0:8080")
}
