package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kameikay/go_sqlc_example/internal/db"
)

func main() {
	ctx := context.Background()
	dbConnection, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/courses")
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()

	queries := db.New(dbConnection)
	// Exemplo de uso do Create
	err = queries.CreateCategory(ctx, db.CreateCategoryParams{
		ID:          uuid.New().String(),
		Name:        "Backend",
		Description: sql.NullString{String: "Backend description"},
	})
	if err != nil {
		panic(err)
	}

	// Exemplo de uso do Update
	err = queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   "25358a0b-8470-4843-b071-b6d0cefdc9c3",
		Name: "Backend Updated",
		Description: sql.NullString{
			String: "Description Updated",
		},
	})
	if err != nil {
		panic(err)
	}

	// Exemplo de uso do Delete
	err = queries.DeleteCategory(ctx, "25358a0b-8470-4843-b071-b6d0cefdc9c3")
	if err != nil {
		panic(err)
	}

	// Exemplo de uso do List All
	categories, err := queries.ListAllCategories(ctx)
	if err != nil {
		panic(err)
	}

	for _, category := range categories {
		fmt.Printf("Category ID: %v | Category Name: %s, | Category Description: %s", category.ID, category.Name, category.Description.String)
	}
}
