package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
)

type Scan struct {
	SCANLOCATION string `json:"scanlocation"`
	IMAGE        []byte `json:"image"`
}

// TODO: Implement concurre cy when model is ready to send picture to model and db at once
func PostScanHandler(c *gin.Context) {

	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the file from the form
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read the file content
	imageBytes, err := ioutil.ReadAll(file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Other fields from JSON payload
	scanLocation := c.Request.FormValue("scanlocation")

	// Create a Scan object
	scan := Scan{
		SCANLOCATION: scanLocation,
		IMAGE:        imageBytes,
	}

	//Using exec to insert data into db
	_, dbErr := database.Db.Exec("INSERT INTO scans(scanlocation, image) values ($1, $2)", scan.SCANLOCATION, scan.IMAGE)
	if dbErr != nil {
		fmt.Println(dbErr)
		c.AbortWithStatusJSON(400, "Could not add scan")
	} else {
		c.JSON(http.StatusOK, "Scan added successfully")
	}

}

func GetScansHandler(c *gin.Context) {
	// Query scans from the database
	rows, err := database.Db.Query("SELECT scanlocation, image FROM scans")
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
		return
	}
	defer rows.Close()

	// Create a slice to store retrieved scans
	var scans []Scan

	// Iterate through the result set
	for rows.Next() {
		var scanLocation string
		var imageBytes []byte

		// Scan values from the result set into variables
		if err := rows.Scan(&scanLocation, &imageBytes); err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
			return
		}

		// Create a Scan object and append it to the scans slice
		scans = append(scans, Scan{SCANLOCATION: scanLocation, IMAGE: imageBytes})
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
		return
	}

	// Marshal scans slice into JSON format
	scansJSON, err := json.Marshal(scans)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal scans"})
		return
	}

	// Return the JSON response containing all the scans
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(scansJSON))

}
