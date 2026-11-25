// La misma API que el endpoint de Serverpod (gs_api/gs_api_server/lib/src/api/api_endpoint.dart), con Gin + GORM,
// contra las MISMAS tablas que crea la migración de Serverpod (autor, tarea con "autorId").
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Autor struct {
	ID     int64   `json:"id" gorm:"primaryKey"`
	Nombre string  `json:"nombre"`
	Tareas []Tarea `json:"tareas,omitempty" gorm:"foreignKey:AutorID"`
}

func (Autor) TableName() string { return "autor" }

type Tarea struct {
	ID      int64  `json:"id" gorm:"primaryKey"`
	Titulo  string `json:"titulo"`
	Hecha   bool   `json:"hecha"`
	AutorID int64  `json:"autorId" gorm:"column:autorId"`
}

func (Tarea) TableName() string { return "tarea" }

type nueva struct {
	Titulo  string `json:"titulo"`
	AutorID int64  `json:"autorId"`
}

func main() {
	dsn := os.Getenv("DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// POOL=n fija el pool de database/sql (máx. abiertas = máx. inactivas = n). Sin POOL, los valores por
	// defecto de Go: abiertas sin límite e inactivas 2. Serverpod usa 10 (maxConnectionCount).
	if p, _ := strconv.Atoi(os.Getenv("POOL")); p > 0 {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(p)
		sqlDB.SetMaxIdleConns(p)
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/tareas/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		var t Tarea
		if err := db.First(&t, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, t)
	})

	// Con Preload: lo que recomienda GORM.
	r.GET("/autores", func(c *gin.Context) {
		limite, _ := strconv.Atoi(c.DefaultQuery("limite", "20"))
		var lista []Autor
		db.Preload("Tareas", func(d *gorm.DB) *gorm.DB { return d.Order("id") }).Order("id").Limit(limite).Find(&lista)
		c.JSON(http.StatusOK, lista)
	})

	// Sin Preload, a mano: una consulta por autor (el N+1).
	r.GET("/autores-n1", func(c *gin.Context) {
		limite, _ := strconv.Atoi(c.DefaultQuery("limite", "20"))
		var lista []Autor
		db.Order("id").Limit(limite).Find(&lista)
		for i := range lista {
			db.Where(`"autorId" = ?`, lista[i].ID).Order("id").Find(&lista[i].Tareas)
		}
		c.JSON(http.StatusOK, lista)
	})

	r.POST("/tareas", func(c *gin.Context) {
		var n nueva
		if err := c.ShouldBindJSON(&n); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		t := Tarea{Titulo: n.Titulo, AutorID: n.AutorID}
		if err := db.Create(&t).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, t.ID)
	})

	r.Run(":8080")
}
