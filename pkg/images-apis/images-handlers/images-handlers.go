package images_handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"time"
)

type Scan struct {
	USERNAME     string `json:"username"`
	SCANLOCATION string `json:"scanlocation"`
	IMAGEURL     string `json:"imageurl"`
	DATECREATED  string `json:"datecreated"`
	TIMECREATED  string `json:"timecreated"`
	IMAGE        []byte `json:"image"`
	PREDICTION   string `json:"prediction"`
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
	updated_at := time.Now()

	// Getting the image
	image, header, imageErr := c.Request.FormFile("image")
	if imageErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": imageErr.Error()})
		return
	}

	// Print out header
	fmt.Println(header)

	// Channel to receive image URL from Google Cloud
	imageURLChan := make(chan string)
	// Channel to receive prediction response
	predictionChan := make(chan string)
	// Channel to receive errors
	errChan := make(chan error, 2)

	// Goroutine to send image to Google Cloud Storage
	go func() {
		imageURL, uploadErr := GoogleCloudUploadHandler(username, scanLocation, image)
		if uploadErr != nil {
			errChan <- uploadErr
			return
		}
		imageURLChan <- imageURL
	}()

	// Goroutine to send image to local endpoint
	go func() {
		// Read the image data
		imageBytes, err := ioutil.ReadAll(image)
		if err != nil {
			errChan <- err
			return
		}

		// Prepare the request to the local endpoint
		req, err := http.NewRequest("POST", "http://127.0.0.1:9000/capstone/api/model", bytes.NewBuffer(imageBytes))
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println(req.Header)
		//req.Header.Set("Content-Type", "multipart/form-data")

		// Send the request
		client := &http.Client{}
		resp, err := client.Post("http://127.0.0.1:9000/capstone/api/model", "image/jpeg", bytes.NewBuffer(imageBytes))
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		// Read the response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		// Assuming the response is in JSON format
		var predictionResponse struct {
			Prediction string `json:"prediction"`
		}
		err = json.Unmarshal(body, &predictionResponse)
		if err != nil {
			errChan <- err
			return
		}

		predictionChan <- predictionResponse.Prediction
	}()

	// Wait for both operations to complete
	var imageURL, prediction string
	for i := 0; i < 2; i++ {
		select {
		case url := <-imageURLChan:
			imageURL = url
		case pred := <-predictionChan:
			prediction = pred
		case err := <-errChan:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Create a Scan object
	scan := Scan{
		USERNAME:     username,
		SCANLOCATION: scanLocation,
		IMAGEURL:     imageURL,
		DATECREATED:  datecreated,
		TIMECREATED:  timecreated,
		PREDICTION:   prediction,
	}

	// Using exec to insert data into db
	_, dbErr := database.Db.Exec("INSERT INTO images(username, scan_location, image_url, date_created, time_created, prediction, updated_at) values ($1, $2, $3, $4, $5, $6, $7)", scan.USERNAME, scan.SCANLOCATION, scan.IMAGEURL, scan.DATECREATED, scan.TIMECREATED, scan.PREDICTION, updated_at.Format("2006-01-02 15:04:05"))
	if dbErr != nil {
		fmt.Println(dbErr)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Could not add scan"})
		return
	}

	// Return the final JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":    "Scan added successfully",
		"prediction": prediction,
	})
}

func GetImagesHandler(c *gin.Context) {
	//TODO: add download scan functionality then return it as image to user
	// Query scans from the database

	//user username
	username := c.Param("username")
	rows, err := database.Db.Query("SELECT username, scan_location,image_url, date_created, time_created, prediction  FROM images WHERE username = $1", username)
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
		var prediction string

		// Scan values from the result set into variables
		if err := rows.Scan(&username, &scanLocation, &imageurl, &datecreated, &timecreated, &prediction); err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
			return
		}

		// Create a Scan object and append it to the scans slice
		scans = append(scans, Scan{USERNAME: username, SCANLOCATION: scanLocation, IMAGEURL: imageurl, DATECREATED: datecreated, TIMECREATED: timecreated, PREDICTION: prediction})

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
