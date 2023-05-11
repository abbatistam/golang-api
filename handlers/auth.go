package handlers

import (
	"context"
	"fmt"
	"log"
	"main/database"
	"main/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const secretKey = "my_secret_key"

func checkUsernameExists(username string) bool {

	// Seleccionar la colección de usuarios en la base de datos
	collection := database.Mg.Db.Collection("users")

	// Crear un filtro que busca un usuario con el mismo nombre de usuario
	filter := bson.M{"username": username}

	// Realizar la búsqueda en la base de datos
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}

	// Si se encontró algún documento, entonces el usuario ya existe en la base de datos
	return count > 0
}

func FindUserByUsername(email string) (*models.Users, error) {

	// Seleccionar la coleccion de usuarios en la base de datos
	collection := database.Mg.Db.Collection("users")

	// Crear un filtro que busca un usuario por el username
	filter := bson.M{"email": email}

	// Declarar la variable que va a contener el usuario que se va a retornar
	var user models.Users
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no user found with username %s", email)
		}
		return nil, fmt.Errorf("error finding user by username: %v", err)
	}

	return &user, nil
}

func Login(c *fiber.Ctx) error {
	// Leer los datos del usuario de la solicitud
	var user models.Users
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Comprobar si existen los datos del usuario en la base de datos
	filter := bson.M{"email": user.Email}
	var dbUser models.Users
	err := database.Mg.Db.Collection("users").FindOne(c.Context(), filter).Decode(&dbUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": 404,
				"message":    "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}

	// Comparar la contraseña proporcionada con la contraseña hash almacenada en la base de datos
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"statusCode": 401,
			"message":    "Passwords do not match",
		})
	}

	// Buscar el perfil del usuario para devolverlo
	var profile models.Profile
	err = database.Mg.Db.Collection("profile").FindOne(c.Context(), bson.M{"user_id": dbUser.ID}).Decode(&profile)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": 404,
			"message":    "Profile not found",
		})
	}

	// Crear un token JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["jti"] = user.Email
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Error generating token",
		})
	}

	//Guardar el token JWT en la base de datos MongoDB
	jwtToken := models.JWTToken{
		Token: signedToken,
	}
	_, err = database.Mg.Db.Collection("jwt").InsertOne(c.Context(), jwtToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}

	// Devolver el token JWT en la respuesta
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": 200,
		"message":    "Login successfull",
		"token":      signedToken,
		"profile":    profile,
	})
}

func Logout(c *fiber.Ctx) error {
	// Obtener el token JWT de la solicitud
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": 400,
			"message":    "Invalid authorization header",
		})
	}
	tokenString := c.Get("Authorization")[7:] // El token JWT está en el header Authorization, después del prefijo "Bearer "

	// Parsear el token JWT y validar la firma
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar que se está usando el algoritmo de firma correcto
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Devolver la clave secreta usada para firmar el token
		return []byte(secretKey), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"statusCode": 401,
			"message":    "Invalid token",
		})
	}

	//Eliminar token JWT de la base de datos
	res, err := database.Mg.Db.Collection("jwt").DeleteOne(c.Context(), bson.M{"token": token})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": 500,
			"message":    "Internal Server Error",
		})
	}
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": 404,
			"message":    "Token not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": 200,
		"message":    "Logout successful",
	})
}
