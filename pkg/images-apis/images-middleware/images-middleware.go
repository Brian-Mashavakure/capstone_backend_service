package images_middleware

import (
	"fmt"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/images-apis/images-handlers"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

type Scan struct {
	USERNAME     string `json:"username"`
	SCANLOCATION string `json:"scanlocation"`
	IMAGEURL     string `json:"imageurl"`
	DATECREATED  string `json:"datecreated"`
	TIMECREATED  string `json:"timecreated"`
	UPDATEDAT    string `json:"updatedat"`
	IMAGE        []byte `json:"image"`
}

func UrlRefreshMiddleware(c *gin.Context) {
	username := c.Request.FormValue("username")
	rows, err := database.Db.Query("SELECT username, scan_location,image_url, date_created, time_created, updated_at FROM images WHERE username = $1", username)
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
		var updatedat string

		// Scan values from the result set into variables
		if err := rows.Scan(&username, &scanLocation, &imageurl, &datecreated, &timecreated, &updatedat); err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
			return
		}

		// Create a Scan object and append it to the scans slice
		scans = append(scans, Scan{USERNAME: username, SCANLOCATION: scanLocation, IMAGEURL: imageurl, DATECREATED: datecreated, TIMECREATED: timecreated, UPDATEDAT: updatedat})
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scans"})
		return
	}

	for _, scan := range scans {
		updatedAt, err := time.Parse(time.RFC3339, scan.UPDATEDAT)
		bucketName := "capstone-project" + scan.USERNAME
		objectName := scan.SCANLOCATION
		if err != nil {
			fmt.Println("Error parsing UPDATEDAT:", err)
			continue
		}
		diff := time.Since(updatedAt)
		if diff.Hours()/24 > 7 {
			newUrl, error := images_handlers.GetSignedUrlHandler(bucketName, objectName)
			if error != nil {
				fmt.Println("New Url Get Failed\n")
			}

			//updating image url
			_, dbErr := database.Db.Exec("INSERT INTO images(updated_at) values ($1) WHERE username = $2", newUrl, username)
			if dbErr != nil {
				fmt.Println(dbErr)
				c.AbortWithStatusJSON(400, "Could not add scan")
				return
			}

			c.JSON(http.StatusOK, "Image URL Updated successfully")
		} else {
			fmt.Printf("The updatedAt of %s is within 7 days.\n", scan.USERNAME)
		}
	}
}
