package dbinfo

import (
	"context"
	"fmt"

	"github.com/Netflix/go-env"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Postgres information needed to form a connection
type PostgresInfo struct {
	URL      string `env:"postgres_url"`
	Port     int32  `env:"postgres_port"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB"`
}

// Construct a Postgres connection url from postgres info
func (pg PostgresInfo) GetUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pg.User, pg.Password, pg.URL, pg.Port, pg.Database)
}

// Redis information
type RedisInfo struct {
	URL  string `env:"redis_url"`
	Port int32  `env:"redis_port"`
}

// Redis URL
func (rd RedisInfo) GetUrl() string {
	return fmt.Sprintf("%s:%d", rd.URL, rd.Port)
}

type Database interface {
	PostgresInfo | RedisInfo
}

func NewDatabaseFromEnv[D Database]() (*D, error) {
	var database D

	//Parse database info
	_, err := env.UnmarshalFromEnviron(&database)

	if err != nil {
		return nil, err
	}

	return &database, nil
}

type User struct {
	Id       uint64 `json:"id"`
	Name     string `json:"username"`
	Password string `json:"password"`
}

// Authenticate the user
func NewUserFromAuth(db *pgx.Conn, username string, password string) (User, error) {
	//Attempt to find user with given password
	user := User{}

	err := db.QueryRow(context.Background(), "select id, name, password from crochet_user where name = $1 and password = $2", username, password).Scan(&user.Id, &user.Name, &user.Password)
	if err != nil {
		fmt.Println()
		return user, fmt.Errorf("user %s does not exist or wrong password was used", username)
	}

	return user, nil
}

type Pattern struct {
	Id      uint64
	Name    string `json:"name"`
	Steps   string `json:"steps"`
	Creator uint64 `json:"creator_id"`
}

type Progress struct {
	Id          uint64
	Progression uint64 `json:"progress"`
	Pattern     uint64 `json:"pattern_id"`
	User        uint64 `json:"user_id"`
}

func JsonError(errar string) gin.H {
	return gin.H{"error": errar}
}
