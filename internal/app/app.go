package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"places/internal/adapter/in"
	"places/internal/adapter/out/foursquare"
	"places/internal/adapter/out/graphhopper"
	"places/internal/adapter/out/openweather"
	"places/internal/service"
	"places/internal/util"
)

type App struct {
	router  *mux.Router
	handler *in.Handler
}

func NewApp() *App {
	util.LoadEnv("config.env")

	// Получаем API ключи из переменных окружения
	graphHopperKey := os.Getenv("GRAPHHOPPER_API_KEY")
	openWeatherKey := os.Getenv("OPENWEATHER_API_KEY")
	foursquareKey := os.Getenv("FOURSQUARE_API_KEY")

	if graphHopperKey == "" || openWeatherKey == "" || foursquareKey == "" {
		log.Fatal("API keys must be set in environment variables")
	}

	// Создаем клиенты
	geocodingClient := graphhopper.NewClient(graphHopperKey)
	weatherClient := openweather.NewClient(openWeatherKey)
	placesClient := foursquare.NewFoursquareClient(foursquareKey)

	// Создаем сервис
	srv := service.NewService(geocodingClient, weatherClient, placesClient)

	// Создаем HTTP handler
	handler := in.NewHandler(srv)

	// Создаем router
	router := mux.NewRouter()

	app := &App{
		router:  router,
		handler: handler,
	}

	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	a.router.HandleFunc("/api/search", a.handler.SearchLocations).Methods("POST")
	a.router.HandleFunc("/api/location/details", a.handler.GetLocationDetails).Methods("POST")

	// Serve static files
	a.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
}

func (a *App) Run(addr string) error {
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, a.router)
}
