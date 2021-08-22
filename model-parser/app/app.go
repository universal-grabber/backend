package app

import (
	"backend/model-parser/lib"
	"backend/model-parser/markers"
	"backend/model-parser/model"
	"crypto/tls"
	"log"
	"net/http"
	"reflect"
)
import "github.com/gin-gonic/gin"

type App struct {
	Addr     string
	CertFile string
	KeyFile  string
	locate   *lib.Locate
}

func (app *App) Run() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	app.locate = new(lib.Locate)

	app.locate.Init()

	r := gin.Default()

	app.routes(r)

	check(r.RunTLS(app.Addr, app.CertFile, app.KeyFile))
}

func (app *App) routes(r *gin.Engine) {
	r.POST("/api/1.0/process", app.process)
	r.POST("/api/1.0/parse", app.parse)

	r.GET("/api/1.0/markers", app.markers)

	r.POST("/api/1.0/process-light", app.processLight)
	r.POST("/api/1.0/parse-light", app.parseLight)
}

func (app *App) processLight(c *gin.Context) {
	p := model.ProcessDataLight{}

	err := c.BindJSON(&p)

	if err != nil {
		c.JSON(400, c.Error(err))
		return
	}

	data, err := app.locate.PrepareProcessData(p)

	if err != nil {
		c.JSON(400, c.Error(err))
		return
	}

	processor := new(lib.Processor)

	result := processor.ProcessData(*data)

	c.String(200, result)
}

func (app *App) parseLight(c *gin.Context) {
	p := model.ProcessDataLight{}

	err := c.BindJSON(&p)

	if err != nil {
		c.JSON(400, c.Error(err))
		return
	}

	data, err := app.locate.PrepareProcessData(p)

	errorType := reflect.TypeOf(err)

	if err != nil && errorType.String() == "*model.Error" && err.(*model.Error).ErrorType == "not-found" {
		result, err := app.staticParse(p)

		if err != nil {
			c.JSON(400, c.Error(err))
			return
		}

		c.JSON(200, result)

		return
	}

	if err != nil {
		c.JSON(400, c.Error(err))
		return
	}

	parser := new(lib.Parser)

	result := parser.Parse(*data)

	c.JSON(200, result)
}

func (app *App) process(c *gin.Context) {
	p := model.ProcessData{}

	err := c.BindJSON(&p)

	if err != nil {
		c.JSON(400, c.Error(err))
	}

	processor := new(lib.Processor)

	result := processor.ProcessData(p)

	c.String(200, result)
}

func (app *App) parse(c *gin.Context) {
	p := model.ProcessData{}

	err := c.BindJSON(&p)

	if err != nil {
		c.JSON(400, c.Error(err))
	}

	parser := new(lib.Parser)

	result := parser.Parse(p)

	c.JSON(200, result)
}

func (app *App) markers(c *gin.Context) {
	type MarkerView struct {
		Name       string                  `json:"name"`
		Parameters []model.MarkerParameter `json:"parameters"`
	}

	var result []MarkerView

	for _, markerType := range markers.GetMarkerTypes() {
		item := MarkerView{}

		item.Name = markerType.GetName()
		item.Parameters = markerType.GetParameters()

		result = append(result, item)
	}

	c.JSON(200, result)
}

func (app *App) staticParse(p model.ProcessDataLight) (*model.Record, error) {
	if *p.Html == "" {
		return nil, model.Error{
			Message:   "html is empty",
			ErrorType: "no-html",
		}
	}

	parser := new(lib.Parser)

	return parser.ParseStaticData(p)
}

func check(err error) {
	if err != nil {
		log.Print(err)
	}
}
