package images_handlers

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
	DATECREATED  string `json:"datecreated"`
	TIMECREATED  string `json:"timecreated"`
	IMAGE        []byte `json:"image"`
}

// TODO: Implement concurrency when model is ready to send picture to model and db at once
func PostImageHandler(c *gin.Context) {
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Other fields from JSON payload
	username := c.Request.FormValue("username")
	scanLocation := c.Request.FormValue("scanlocation")
	datecreated := c.Request.FormValue("datecreated")
	timecreated := c.Request.FormValue("timecreated")

	//getting the image
	image, header, imageErr := c.Request.FormFile("image")
	if imageErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	//print out header
	fmt.Println(header)

	//sending image to cloud storage
	imageURL, uploadErr := GoogleCloudUploadHandler(username, scanLocation, image)
	if err != nil {
		fmt.Printf("Cloud storage upload failed: %v", uploadErr)
	}

	// Create a Scan object
	scan := Scan{
		USERNAME:     username,
		SCANLOCATION: scanLocation,
		IMAGEURL:     imageURL,
		DATECREATED:  datecreated,
		TIMECREATED:  timecreated,
	}

	//TODO: find a fix for image going to cloud store but potentially not going to db

	//Using exec to insert data into db
	_, dbErr := database.Db.Exec("INSERT INTO images(username, scan_location, image_url, date_created, time_created) values ($1, $2, $3, $4, $5)", scan.USERNAME, scan.SCANLOCATION, scan.IMAGEURL, scan.DATECREATED, scan.TIMECREATED)
	if dbErr != nil {
		fmt.Println(dbErr)
		c.AbortWithStatusJSON(400, "Could not add scan")
		return
	}

	c.JSON(http.StatusOK, "Scan added successfully")
}

func GetImagesHandler(c *gin.Context) {
	//TODO: add download scan functionality then return it as image to user
	// Query scans from the database

	//user username
	username := c.Param("username")
	rows, err := database.Db.Query("SELECT username, scan_location,image_url, date_created, time_created FROM images WHERE username = $1", username)
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
		var username string
		var imageurl string
		var datecreated string
		var timecreated string

		// Scan values from the result set into variables
		if err := rows.Scan(&username, &scanLocation, &imageurl, &datecreated, &timecreated); err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
			return
		}

		// Create a Scan object and append it to the scans slice
		scans = append(scans, Scan{USERNAME: username, SCANLOCATION: scanLocation, IMAGEURL: imageurl, DATECREATED: datecreated})

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
