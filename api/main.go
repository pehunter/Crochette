package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type subscription struct {
	Key string `json:"key"`
}

type light struct {
	Name       string  `json:"name"`
	Brightness float32 `json:"brightness"`
	R          uint8   `json:"r"`
	G          uint8   `json:"g"`
	B          uint8   `json:"b"`
	Service    string  `json:"service"`
	ID         uint8   `json:"id"`
}

// Update light with incoming state
func update(c *gin.Context) {
	//Get JSON
	var light light

	if err := c.BindJSON(&light); err != nil {
		return
	}

	//Find the appropriate server

	//Attempt to update light on that server

	//Return status
	c.IndentedJSON(http.StatusCreated, light)
}

func main() {
	router := gin.Default()
	router.POST("/update", update)

	router.Run("localhost:8080")
}
