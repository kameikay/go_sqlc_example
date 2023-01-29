* Usado para POSTGRES, MYSQL e SQLITE

### Passo-a-passo

1. Cria um arquivo yaml

```yaml
version: "2"
sql:
  - schema: "sql/migrations"
  queries: "sql/queries"
  engine: "mysql"
  gen:
    go:
      package: "db"
      out: "internal/db"
```

1. Considerando que está sendo usando [Migrations](https://www.notion.so/90abe36b6cd14614aee4b32dd8b36c73)
2. Na pasta sql/queries, criar um arquivo query.sql. Este arquivo terá as queries com anotations:

```sql
-- name: ListAllCategories :many
SELECT * FROM categories;
```

1. Depois, executar o comando:

```bash
sqlc generate
```

Este comando cria (dentro da pasta sql/migrations, que está setada no arquivo yaml), três arquivos: db.go, models.go e query.sql.go.

1. Em query.sql serão criadas as queries:

```sql
-- name: ListAllCategories :many
SELECT * FROM categories;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = ?;

-- name: CreateCategory :exec
INSERT INTO categories (id, name, description) VALUES (?, ?, ?)

-- name: UpdateCategory :exec
UPDATE categories SET name = ?, description = ? WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;
```

1. Para inicializar, dentro de cmd/run_sqlc/main.go colocamos:

```go
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
	ctx := context.Background() // Cria um contexto
	// Cria conexão com o banco de dados
	dbConnection, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/courses")
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()
	// pega o db que foi criado pelo SQLC em internal
	queries := db.New(dbConnection)

	// Exemplo de criação da categoria
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

	// Exemplo de uso de Delete
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
		fmt.Printf("Category ID: %v | Category Name: %s, | Category Description: %v", category.ID, category.Name, category.Description.String)
	}
}
```

---

### Transactions

Utilizado para quando executar alguma query, caso dê certo, faça o commit, caso dê errado, faça o rollback.

1. Cria-se uma struct:

```go
type CourseDB struct {
	dbConnection *sql.DB
	*db.Queries
}
```

1. Depois, cria-se uma função de criação:

```go
func NewCourseDB(dbConnection *sql.DB) *CourseDB {
	return &CourseDB{
		dbConnection: dbConnection,
		Queries: db.New(dbConnection),
	}
}
```

1. Cria-se uma função atrelada à esta struct, a qual será a chamada à transaction. Esta função receberá um context e uma outra função. Caso dê algum erro para executar esta função, fará o rollback e retornará erro. Caso dê certo, dará o commit:

```go
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
```

1. Um exemplo prático: precisamos criar uma categoria e, logo em seguida, um curso. Desta forma, já chamaremos a função callTx que irá verificar se der algum erro e dar rollback ou commit:

```go
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
```

> Tip: sempre que o SQLC ver um float, por default, ele vai gerar o model como string. Para mudar isso, deve fazer um override no sqlc.yaml:
> 

```go
version: "2"
sql:
- schema: "sql/migrations"
  queries: "sql/queries"
  engine: "mysql"
  gen:
    go:
      package: "db"
      out: "internal/db"
      overrides:
      - db_type: "decimal"
        go_type: "float64"
```

---

### Usando Join

1. Para trazer o curso e o nome da categoria (e não somente o ID), cria-se uma nova query no query.sql:

```sql
-- name: ListCourses :many
SELECT c.*, ca.name as category_name
FROM courses c 
JOIN categories ca
ON c.category_id = ca.id;
```

1. Executa a função gerada pelo SQLC:

```go
courses, err := queries.ListCourses(ctx)
	if err != nil {
		panic(err)
	}

	for _, course := range courses {
		fmt.Printf("Category: %s, Course ID: %s, Course Name: %s, Course Description: %s, Course Price: %f", course.CategoryName, course.ID, course.Name, course.Description.String, course.Price)
	}
```