package app

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"places/internal/adapter/in"
	"places/internal/adapter/out/foursquare"
	"places/internal/adapter/out/graphhopper"
	"places/internal/adapter/out/openweather"
	"places/internal/service"
)

type App struct {
	router  *mux.Router
	handler *in.Handler
}

func loadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // пропускаем пустые строки и комментарии
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value) // устанавливаем переменную окружения
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func NewApp() *App {
	loadEnv("config.env")

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
	service := service.NewService(geocodingClient, weatherClient, placesClient)

	// Создаем HTTP handler
	handler := in.NewHandler(service)

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
	// Настройка CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(a.router)

	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, handler)
}
