package nba

import (
	"encoding/json"
	"fmt"
	"forwardproxy/client/nbaclient"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestPayload represents the JSON payload structure
type RequestPayload struct {
	Endpoint string            `json:"endpoint" binding:"required"`
	Params   map[string]string `json:"params"`
}

// HandleNBA handles POST requests to /nba
func HandleNBA(c *gin.Context) {
	var payload RequestPayload

	// Bind JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Create NBA client and make GET request
	nbaClient := nbaclient.NewNBAClient()
	fmt.Println(payload.Endpoint, payload.Params)
	data, err := nbaClient.GetNBAData(payload.Endpoint, payload.Params, nil)
	if err != nil {
		log.Printf("Error making NBA API request: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch data from NBA API"})
		return
	}
	fmt.Println(string(data))
	// Parse JSON response to return as structured JSON
	var jsonResponse interface{}
	if err := json.Unmarshal(data, &jsonResponse); err != nil {
		log.Printf("Error parsing NBA API response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	// Return the NBA API response
	c.JSON(http.StatusOK, jsonResponse)
}
