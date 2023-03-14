package handlers

import (
	"main/database"
	"main/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllProducts(c *fiber.Ctx) error {
	query := bson.D{{}}
	cursor, err := database.Mg.Db.Collection("Products").Find(c.Context(), query)
	if err != nil {
		//return c.Status(500).SendString(err.Error())
		e := models.Error{Message: err.Error(), StatusCode: 500}
		return c.JSON(e)
	}

	var products []models.Product = make([]models.Product, 0)

	if err := cursor.All(c.Context(), &products); err != nil {
		//return c.Status(500).SendString(err.Error())
		e := models.Error{Message: err.Error(), StatusCode: 500}
		return c.JSON(e)
	}

	response := models.ProductResponse{Items: products, Total: len(products)}

	return c.JSON(response)
}

func NewProduct(c *fiber.Ctx) error {
	collection := database.Mg.Db.Collection("Products")

	product := new(models.Product)

	if err := c.BodyParser(product); err != nil {
		//return c.Status(400).SendString(err.Error())
		e := models.Error{Message: err.Error(), StatusCode: 400}
		return c.JSON(e)
	}

	product.ID = ""

	insertionResult, err := collection.InsertOne(c.Context(), product)
	if err != nil {
		//return c.Status(500).SendString(err.Error())
		e := models.Error{Message: err.Error(), StatusCode: 500}
		return c.JSON(e)
	}

	filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
	createdRecord := collection.FindOne(c.Context(), filter)

	createdProduct := &models.Product{}
	createdRecord.Decode(createdProduct)

	return c.Status(201).JSON(createdProduct)
}

func EditProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := primitive.ObjectIDFromHex(idParam)

	if err != nil {
		//return c.SendStatus(400)
		e := models.Error{Message: err.Error(), StatusCode: 400}
		return c.JSON(e)
	}

	product := new(models.Product)

	if err := c.BodyParser(product); err != nil {
		//return c.Status(400).SendString(err.Error())
		e := models.Error{Message: err.Error(), StatusCode: 400}
		return c.JSON(e)
	}

	query := bson.D{{Key: "_id", Value: productID}}
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "name", Value: product.Name},
				{Key: "category", Value: product.Category},
				{Key: "image", Value: product.Image},
				{Key: "description", Value: product.Description},
				{Key: "price", Value: product.Price},
			},
		},
	}
	err = database.Mg.Db.Collection("Products").FindOneAndUpdate(c.Context(), query, update).Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			//return c.SendStatus(404)
			e := models.Error{Message: "Not Found", StatusCode: 404}
			return c.JSON(e)
		}
		//return c.SendStatus(500)
		e := models.Error{Message: err.Error(), StatusCode: 500}
		return c.JSON(e)
	}

	product.ID = idParam
	return c.Status(200).JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {
	noteID, err := primitive.ObjectIDFromHex(
		c.Params("id"),
	)

	if err != nil {
		//return c.SendStatus(400)
		//return c.SendFile()
		e := models.Error{Message: err.Error(), StatusCode: 400}
		return c.JSON(e)
	}

	query := bson.D{{Key: "_id", Value: noteID}}
	result, err := database.Mg.Db.Collection("Products").DeleteOne(c.Context(), &query)

	if err != nil {
		e := models.Error{Message: err.Error(), StatusCode: 500}
		//return c.SendStatus(500)
		return c.JSON(e)
	}

	if result.DeletedCount < 1 {
		e := models.Error{Message: "Not Found", StatusCode: 404}
		//return c.SendStatus(404)
		return c.JSON(e)
	}

	return c.JSON(query[0].Value)
	//return c.SendStatus(204)
}
