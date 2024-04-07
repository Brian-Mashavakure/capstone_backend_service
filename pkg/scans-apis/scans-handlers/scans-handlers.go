package scans_handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
)

type Scan struct {
	USERNAME     string `json:"username"`
	SCANLOCATION string `json:"scanlocation"`
	IMAGEURL     string `json:"imageurl"`
}

// TODO: Implement concurrency when model is ready to send picture to model and db at once
func PostScanHandler(c *gin.Context) {
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Other fields from JSON payload
	username := c.Request.FormValue("username")
	scanLocation := c.Request.FormValue("scanlocation")

	//getting the image
	image, header, imageErr := c.Request.FormFile("image")
	if imageErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	//print out header
	fmt.Println(header)

	//sending image to cloud storage
	imageURL, uploadErr := GoogleCloudHandler(username, scanLocation, image)
	if err != nil {
		fmt.Printf("Cloud storage upload failed: %v", uploadErr)
	}

	// Create a Scan object
	scan := Scan{
		USERNAME:     username,
		SCANLOCATION: scanLocation,
		IMAGEURL:     imageURL,
	}

	//TODO: find a fix for image going to cloud store but potentially not going to db

	//Using exec to insert data into db
	_, dbErr := database.Db.Exec("INSERT INTO scanstable(username, scanlocation, imageurl) values ($1, $2, $3)", scan.USERNAME, scan.SCANLOCATION, scan.IMAGEURL)
	if dbErr != nil {
		fmt.Println(dbErr)
		c.AbortWithStatusJSON(400, "Could not add scan")
		return
	}

	c.JSON(http.StatusOK, "Scan added successfully")
}

func GetScansHandler(c *gin.Context) {
	//TODO: add download scan functionality then return it as image to user
	// Query scans from the database
	rows, err := database.Db.Query("SELECT scanlocation, imageurl FROM scanstable")
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
		var imageURL string

		// Scan values from the result set into variables
		if err := rows.Scan(&scanLocation, &imageURL); err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
			return
		}

		// Create a Scan object and append it to the scans slice
		scans = append(scans, Scan{SCANLOCATION: scanLocation, IMAGEURL: imageURL})
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
