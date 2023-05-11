package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"main/database"
	"main/models"
)

func UpdateProfile(c *fiber.Ctx) error {
	// Obtenemos el valor del userID del query usando fiber
	userID := c.Params("id")

	// Transformamos userID a ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid userID",
		})
	}

	// Parseamos el body de la petición a un struct Profile
	updateProfile := new(models.Profile)
	if err := c.BodyParser(updateProfile); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid request body",
		})
	}

	filter := bson.M{"_id": objID}

	count, err := database.Mg.Db.Collection("users").CountDocuments(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": 404,
			"message":    "User not found",
		})
	}

	update := bson.M{
		"first_name": updateProfile.FirstName,
		"last_name":  updateProfile.LastName,
		"phone":      updateProfile.Phone,
	}

	updateNotNull := bson.M{}
	for key, value := range update {
		if value != "" {
			updateNotNull[key] = value
		}
	}

	filterProfile := bson.M{"user_id": objID}

	_, err = database.Mg.Db.Collection("profile").UpdateOne(c.Context(), filterProfile, bson.M{"$set": updateNotNull})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(updateNotNull)

}
