package main

import (
	"database/sql"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"

	"gourmeg-htmx/api"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

    postgres_uri := os.Getenv("POSTGRESQL_URI")
    if len(postgres_uri) == 0 {
        postgres_uri =  "host=localhost user=root password=secret dbname=gourmeg_2 sslmode=disable"
    }

	db, err := sql.Open("postgres", postgres_uri)
	if err != nil {
		e.Logger.Fatalf("unable to open database connection: %b", err)
	}
	if err := db.Ping(); err != nil {
		e.Logger.Fatalf("unable to connect to database %b", err)
	}
    db.SetMaxIdleConns(20)
    db.SetConnMaxLifetime(1000 * time.Millisecond)
	h := &api.Handler{DB: db}
	defer db.Close()

    templates_dir := os.Getenv("TEMPLATES_DIR")
    if len(templates_dir) == 0 {
        templates_dir = "templates/*/*.html"
    }

	t, err := template.ParseGlob(templates_dir)
	if err != nil {
		e.Logger.Fatalf("unable to parse templates: %b", err)
	}
	e.Renderer = &Templates{templates: t}

	e.Static("/css", "public/css")
	e.Static("/js", "public/js")
	e.Static("/static", "public/assets")

	// pages
	e.GET("/", h.Index)
	e.GET("/add", h.Add)
	e.GET("/search", h.Search)

	// recipe
	e.GET("/recipe/:id", h.GetRecipe)
	e.POST("/recipe", h.PostRecipe)
    e.PUT("/recipe/refetch/:id", h.RefetchRecipe)

	// list
	e.GET("/list/:id", h.GetList)
	e.GET("/list/:id/edit", h.EditList)
	e.DELETE("/list/:id", h.DeleteList)
	e.POST("/list", h.PostList)
	e.POST("/list_search", h.GetLists)

	// list item
	e.GET("/item/:id", h.GetItem)
	e.PUT("/item/:type/:id", h.PutItem)
	e.GET("/item/:id/edit", h.EditItem)
	e.GET("/item/recipe/add", h.AddRecipeItem)
	e.GET("/item/list/add", h.AddListItem)
	e.DELETE("/item/:type/:id", h.DeleteItem)
    e.POST("item/search", h.ItemSearch)

	e.Debug = true
	e.Logger.Fatal(e.Start(":1323"))
}
