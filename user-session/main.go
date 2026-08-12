package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Netflix/go-env"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type PostgresInfo struct {
	URL      string `env:"postgres_url"`
	Port     int32  `env:"postgres_port"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB"`
}

func (pg PostgresInfo) toUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pg.User, pg.Password, pg.URL, pg.Port, pg.Database)
}

type user struct {
	Id       uint64
	Name     string `json:name`
	Password string `json:password`
}

type session struct {
	SessionID  string
	UserID     uint64
	LastUpdate time.Time
}

type SessionMap struct {
	data map[string]session
	lock sync.Mutex
}

type ActiveUserMap struct {
	data map[uint64]string
	lock sync.Mutex
}

// User ID -> Session
var activeUsers ActiveUserMap

// Session ID -> Session
var sessions SessionMap

// Get a fresh, new session key
var sessionCount uint64

func getSessionKey() string {
	sessionCount += 1
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
			c.String(http.StatusBadRequest, "Request has missing parameters.")
			return
		}

		//Authenticate user
		user, err := authUser(db, user.Name, user.Password)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		//Lock activeUsers and defer unlock
		activeUsers.lock.Lock()
		defer activeUsers.lock.Unlock()

		//Check if the user already has a sesson. If so, stop
		if _, ok := activeUsers.data[user.Id]; ok == true {
			c.String(http.StatusUnauthorized, "User has already logged in.")
			return
		}

		//Acquire session lock
		sessions.lock.Lock()
		defer sessions.lock.Unlock()

		//Insert them into the session
		sessionKey := getSessionKey()
		sessions.data[sessionKey] = session{
			UserID:     user.Id,
			SessionID:  sessionKey,
			LastUpdate: time.Now(),
		}
		activeUsers.data[user.Id] = sessionKey

		//Return status
		c.String(http.StatusOK, sessionKey)
	}
}

// Get the user associated with the session
func sessionUser(c *gin.Context) {
	sessionKey, ok := c.Params.Get("session")
	if !ok {
		c.String(http.StatusBadRequest, "No session key provided.")
		return
	}

	//Lock sessions
	sessions.lock.Lock()
	defer sessions.lock.Unlock()

	if session, ok := sessions.data[sessionKey]; !ok {
		c.String(http.StatusNotFound, "The session key has expired.")
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", session.LastUpdate.String())

		//Refresh session end time
		session.LastUpdate = time.Now()
		sessions.data[sessionKey] = session

		//Return user ID
		c.String(http.StatusOK, strconv.FormatUint(session.UserID, 10))
	}

}

// How long a session will last
var sessionTimeout time.Duration

// Remove old sessions
func clearOldSessions() {
	//Lock sessions
	sessions.lock.Lock()
	defer sessions.lock.Unlock()

	//Iterate sessions and find ones < timeout
	expiredSessions := make([]string, len(sessions.data))
	for sessionKey, session := range sessions.data {
		timeDiff := time.Since(session.LastUpdate)

		//Session timeout reached
		if timeDiff > sessionTimeout {
			expiredSessions = append(expiredSessions, sessionKey)
		}
	}

	//Return now if all sessions are valid
	if len(expiredSessions) <= 0 {
		return
	}

	activeUsers.lock.Lock()
	defer activeUsers.lock.Unlock()

	//Remove expired keys
	//From sessions AND users
	for _, expiredKey := range expiredSessions {
		session := sessions.data[expiredKey]
		fmt.Printf("Deleting %s", expiredKey)
		//Remove from sessions
		delete(sessions.data, expiredKey)

		//Remove user from activeUsers
		delete(activeUsers.data, session.UserID)
	}
}

func main() {
	//Parse Postgres info
	var postgres PostgresInfo

	_, err := env.UnmarshalFromEnviron(&postgres)

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	//Make sessions/activeUsers map
	sessions = SessionMap{data: make(map[string]session)}
	activeUsers = ActiveUserMap{data: make(map[uint64]string)}
	sessionCount = 0
	sessionTimeout = time.Second * 18

	//Connect to Postgres
	conn, err := pgx.Connect(context.Background(), postgres.toUrl())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to postgres :( %s", err)
		return
	}
	defer conn.Close(context.Background())

	//Start session garbage collection
	go func() {
		for {
			clearOldSessions()
			time.Sleep(time.Second * 5)
		}
	}()

	router := gin.Default()
	router.POST("/login", addUser(conn))
	router.GET("/:session", sessionUser)

	router.Run("0.0.0.0:8080")
}
