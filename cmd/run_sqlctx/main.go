package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kameikay/go_sqlc_example/internal/db"
)

type CourseDB struct {
	dbConnection *sql.DB
	*db.Queries
}

type CourseParams struct {
	ID          string
	Name        string
	Description sql.NullString
	Price       float64
}

type CategoryParams struct {
	ID          string
	Name        string
	Description sql.NullString
}

func NewCourseDB(dbConnection *sql.DB) *CourseDB {
	return &CourseDB{
		dbConnection: dbConnection,
		Queries:      db.New(dbConnection),
	}
}

func (c *CourseDB) callTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := c.dbConnection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query := db.New(tx)
	err = fn(query)
	if err != nil {
		if errRb := tx.Rollback(); errRb != nil {
			return fmt.Errorf("error on rollback: %v, original error: %w", errRb, err)
		}
		return err
	}

	return tx.Commit()
}

func (c *CourseDB) CreateCourseAndCategory(ctx context.Context, argsCategory CategoryParams, argsCourse CourseParams) error {
	err := c.callTx(ctx, func(q *db.Queries) error {
		var err error
		err = q.CreateCategory(ctx, db.CreateCategoryParams{
			ID:          argsCategory.ID,
			Name:        argsCategory.Name,
			Description: argsCategory.Description,
		})
		if err != nil {
			return err
		}

		err = q.CreateCourse(ctx, db.CreateCourseParams{
			ID:          argsCourse.ID,
			Name:        argsCourse.Name,
			Description: argsCourse.Description,
			CategoryID:  argsCategory.ID,
			Price:       argsCourse.Price,
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()
	dbConnection, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/courses")
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()

	queries := db.New(dbConnection)

	// courseArgs := CourseParams{
	// 	ID:          uuid.New().String(),
	// 	Name:        "Go",
	// 	Description: sql.NullString{String: "Go Course", Valid: true},
	// 	Price:       123.1,
	// }
	// categoryArgs := CategoryParams{
	// 	ID:          uuid.New().String(),
	// 	Name:        "Backend",
	// 	Description: sql.NullString{String: "Backend Courses", Valid: true},
	// }

	// courseDB := NewCourseDB(dbConnection)
	// err = courseDB.CreateCourseAndCategory(ctx, categoryArgs, courseArgs)
	// if err != nil {
	// 	panic(err)
	// }

	courses, err := queries.ListCourses(ctx)
	if err != nil {
		panic(err)
	}

	for _, course := range courses {
		fmt.Printf("Category: %s, Course ID: %s, Course Name: %s, Course Description: %s, Course Price: %f", course.CategoryName, course.ID, course.Name, course.Description.String, course.Price)
	}
}
