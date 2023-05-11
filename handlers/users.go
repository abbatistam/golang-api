package handlers

import (
	"context"
	"log"
	"main/database"
	"main/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func checkEmailExists(email string) bool {

	// Seleccionar la colección de usuarios en la base de datos
	collection := database.Mg.Db.Collection("users")

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

// CreateUser crea un nuevo usuario
// @Summary Crear un nuevo usuario
// @Description Crea un nuevo usuario y su perfil
// @Tags Usuario
// @Accept json
// @Produce json
// @Param name body string true "Nombre del usuario"
// @Param email body string true "Correo electrónico del usuario"
// @Param password body string true "Contraseña del usuario"
// @Param role body string true "Rol del usuario"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func CreateUser(c *fiber.Ctx) error {

	collection := database.Mg.Db.Collection("users")
	profileCollection := database.Mg.Db.Collection("profile")

	user := new(models.Users)

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid request body",
		})
	}

	// Check if email exists
	if checkEmailExists(user.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Email already exists",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Error hashing password",
		})
	}

	// Insert user
	res, err := collection.InsertOne(c.Context(), bson.M{
		"name":     user.Name,
		"email":    user.Email,
		"password": hashedPassword,
		"role":     user.Role,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal server error",
		})
	}

	// Convert inserted ID to ObjectID
	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Cannot convert inserted ID to ObjectID",
		})
	}

	// Insert profile
	profile := models.Profile{
		UserID: insertedID,
	}

	res2, err := profileCollection.InsertOne(c.Context(), bson.M{
		"user_id": res.InsertedID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal server error",
		})
	}

	insertedUserID, ok := res2.InsertedID.(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Cannot convert inserted ID to ObjectID",
		})
	}

	profile.ID = insertedUserID

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      res.InsertedID,
		"name":    user.Name,
		"email":   user.Email,
		"role":    user.Role,
		"profile": profile,
	})

}

// @Summary Get Users
// @Description Obtiene una lista de usuarios con sus perfiles.
// @Tags Users
// @Accept json
// @Produce json
// @Param search query string false "Parámetro opcional para buscar usuarios por nombre."
// @Success 200 {object} GetUsersResponse
// @Router /users [get]
func GetUsers(c *fiber.Ctx) error {

	query := bson.D{{}}

	if len(c.Query("search")) > 0 {
		query = bson.D{{Key: "name", Value: primitive.Regex{Pattern: c.Query("search"), Options: "i"}}}
	}

	// Get all fields except password
	projection := bson.M{"password": 0}

	cursor, err := database.Mg.Db.Collection("users").Find(c.Context(), query, options.Find().SetProjection(projection))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal server error asd1",
			"error":      err,
		})
	}
	defer cursor.Close(context.Background())

	var usersWithProfile []map[string]interface{}

	for cursor.Next(context.Background()) {
		var user models.Users
		if err := cursor.Decode(&user); err != nil {
			return err
		}
		var profile models.Profile
		err = database.Mg.Db.Collection("profile").FindOne(c.Context(), bson.M{"user_id": user.ID}).Decode(&profile)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": 404,
				"message":    "User not found",
			})
		}
		//// this way

		userWithProfile := make(map[string]interface{})
		userWithProfile["id"] = user.ID
		userWithProfile["name"] = user.Name
		userWithProfile["email"] = user.Email
		userWithProfile["role"] = user.Role
		userWithProfile["profile"] = profile
		usersWithProfile = append(usersWithProfile, userWithProfile)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"items": usersWithProfile,
		"total": len(usersWithProfile),
	})
}

// @Summary Get User
// @Description Obtiene un usuario por su ID.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "El ID del usuario."
// @Success 200 {object} models.Users
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /users/{id} [get]
func GetUser(c *fiber.Ctx) error {
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

	var user models.Users
	err = database.Mg.Db.Collection("users").FindOne(c.Context(), bson.M{"_id": objectID}, options.FindOne().SetProjection(projection)).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": 404,
			"message":    "User not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&user)
}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid ID",
		})
	}
	user := new(models.Users)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid request body",
		})
	}
	filter := bson.M{"_id": objectId}

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

	// Crear un mapa con los campos actualizables
	update := bson.M{
		"name":     user.Name,
		"email":    user.Email,
		"password": user.Password,
	}

	// Crear un bson.M con los datos no nulos
	updateNotNull := bson.M{}
	for key, value := range update {
		if value != "" {
			updateNotNull[key] = value
		}
	}

	_, err = database.Mg.Db.Collection("users").UpdateOne(c.Context(), filter, bson.M{"$set": updateNotNull})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(&user)
}

/*
func deleteUser(c *fiber.Ctx) error {
    id := c.Params(\"id")
    _, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
    if err != nil {
        return err
    }
    return c.SendStatus(http.StatusOK)
}
*/
