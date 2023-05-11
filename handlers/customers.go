package handlers

import (
	"context"
	"main/database"
	"main/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

/*
func checkEmailExists(email string) bool {

		// Seleccionar la colección de usuarios en la base de datos
		collection := database.Mg.Db.Collection("customers")

		// Crear un filtro que busca un usuario con el mismo nombre de usuario
		filter := bson.M{"email": email}

		// Realizar la búsqueda en la base de datos
		count, err := collection.CountDocuments(context.Background(), filter)
		if err != nil {
			log.Fatal(err)
		}

		// Si se encontró algún documento, entonces el usuario ya existe en la base de datos
		return count > 0
	}
*/
func CreateCustomer(c *fiber.Ctx) error {

	collection := database.Mg.Db.Collection("customers")

	customer := new(models.Customer)

	if err := c.BodyParser(&customer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid request body",
		})
	}

	// Check if email exists
	if checkEmailExists(customer.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Email already exists",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(customer.Password), 8)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Error hashing password",
		})
	}

	// Insert customer
	res, err := collection.InsertOne(c.Context(), bson.M{
		"name":         customer.Name,
		"email":        customer.Email,
		"password":     hashedPassword,
		"phone":        customer.Phone,
		"affiliate_id": customer.AffiliateID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           res.InsertedID,
		"name":         customer.Name,
		"email":        customer.Email,
		"phone":        customer.Phone,
		"affiliate_id": customer.AffiliateID,
	})

}

func GetCustomers(c *fiber.Ctx) error {

	//Creating a new context with a timout of 10 seconds
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	//Extracting search query parameter, and initializing the query
	search := c.Query("search")
	query := bson.D{{}}

	if len(c.Query("search")) > 0 {
		//Setting query field to name and value to a regex object that matches the search query with options to ignore case sensitivity
		query = bson.D{{Key: "name", Value: primitive.Regex{Pattern: search, Options: "i"}}}
	}

	//Setting projection to exclude password field
	projection := bson.M{"password": 0}

	//Getting all customers that match the search query and projecting all fields except password
	cursor, err := database.Mg.Db.Collection("customers").Find(ctx, query, options.Find().SetProjection(projection))

	if err != nil {
		//Returning an error response if an error occurs
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Unable to retrieve customers",
		})
	}

	//Closing cursor once all customers have been returned
	defer cursor.Close(ctx)

	var customers []bson.M

	if err := cursor.All(ctx, &customers); err != nil {
		//Returning an error response if an error occurs
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Unable to retrieve customers",
		})
	}

	//Returning a success response with customers and total number of customers found
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"items": customers,
		"total": len(customers),
	})
}

func GetCustomer(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid ID",
		})
	}

	// Get all fields except password
	projection := bson.M{"password": 0}

	var customer models.Customer
	err = database.Mg.Db.Collection("customers").FindOne(ctx, bson.M{"_id": objectID}, options.FindOne().SetProjection(projection)).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": 404,
				"message":    "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // En caso contrario, se retorna un error interno del servidor.
			"statusCode": 500,
			"message":    "Internal Server Error",
			"error":      err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(&customer)
}

func UpdateCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid ID",
		})
	}
	customer := new(models.Customer)
	if err := c.BodyParser(customer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid request body",
		})
	}
	filter := bson.M{"_id": objectId}

	count, err := database.Mg.Db.Collection("customers").CountDocuments(c.Context(), filter)
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

	// Crear un mapa con los campos actualizables
	update := bson.M{
		"name":     customer.Name,
		"email":    customer.Email,
		"password": customer.Password,
		"phone":    customer.Phone,
	}

	// Crear un bson.M con los datos no nulos
	updateNotNull := bson.M{}
	for key, value := range update {
		if value != "" {
			updateNotNull[key] = value
		}
	}

	_, err = database.Mg.Db.Collection("customers").UpdateOne(c.Context(), filter, bson.M{"$set": updateNotNull})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(&customer)
}

func DeleteCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid ID",
		})
	}
	result, err := database.Mg.Db.Collection("customers").DeleteOne(c.Context(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": 404,
			"message":    "Customer not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": 200,
		"message":    "Customer deleted successfully",
		"id":         id,
	})
}
