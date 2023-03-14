package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func UploadMultiFiles(c *fiber.Ctx) error {
	// Parse the multipart form:
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	// => *multipart.Form

	// Get all files from "files" key:
	files := form.File["files"]
	// => []*multipart.FileHeader

	// Loop through files:
	for _, file := range files {
		fmt.Println(file.Filename, file.Size, file.Header["Content-Type"][0])

		// Save the files to disk:
		err := c.SaveFile(file, fmt.Sprintf("./imgs/%s", file.Filename))

		// Check for errors
		if err != nil {
			return err
		}

		// Save the url
		url := fmt.Sprintf("/imgs/%s", file.Filename)

		// Add the url to headers file
		file.Header.Add("url", url)
	}
	return c.JSON(files)
}