package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// --- KONFIGURASI & VARIABEL GLOBAL ---

var openWeatherAPIKey string
var router *gin.Engine

// Konstanta untuk URL API OpenWeatherMap agar mudah dikelola.
const (
	geoAPIURL          = "https://api.openweathermap.org/geo/1.0/direct"
	forecastAPIURL     = "https://api.openweathermap.org/data/2.5/forecast"
	airPollutionAPIURL = "https://api.openweathermap.org/data/2.5/air_pollution"
)

// --- STRUKTUR DATA (STRUCTS) ---

type CleanedCity struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Country     string  `json:"country"`
}

type geoResult struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	State   string  `json:"state,omitempty"`
}

type WeatherResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   float64 `json:"deg"`
		} `json:"wind"`
		Pop float64 `json:"pop"`
	} `json:"list"`
	City struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"city"`
}

type AirPollutionResponse struct {
	List []struct {
		Main struct {
			Aqi int `json:"aqi"`
		} `json:"main"`
		Components struct {
			Co    float64 `json:"co"`
			No2   float64 `json:"no2"`
			O3    float64 `json:"o3"`
			So2   float64 `json:"so2"`
		} `json:"components"`
	} `json:"list"`
}


// --- FUNGSI INIT & HANDLER UNTUK VERCEL ---

// setupRouter menginisialisasi Gin router
func setupRouter() *gin.Engine {
	// Memuat environment variable (untuk pengembangan lokal)
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan. Mengandalkan Environment Variables Vercel.")
	}
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
    r.Use(gin.Logger(), gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true 
	r.Use(cors.New(config))

	api := r.Group("/api")
	{
		api.GET("/search", searchCitiesHandler)
		api.GET("/weather", getWeatherHandler)
		api.GET("/air-pollution", getAirPollutionHandler)
	}
	
	r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Go API (handler package) is running on Vercel"})
    })
    
    return r
}

// init() menjalankan setupRouter HANYA SEKALI saat fungsi dimuat (cold start).
func init() {
    router = setupRouter()
}

// MainHandler adalah fungsi yang diekspor Vercel untuk menjalankan Serverless Function.
// Nama ini dipilih karena "Handler" mungkin berbenturan dengan variabel Gin sebelumnya.
func MainHandler(w http.ResponseWriter, r *http.Request) {
    // Arahkan request ke instance Gin router
    router.ServeHTTP(w, r)
}


// --- FUNGSI UTAMA LOKAL (UNTUK TESTING LOKAL) ---

// Fungsi main() harus tetap dipertahankan dengan package main untuk kompilasi lokal.
// Namun, karena kita mengubah package ke 'handler', kita harus membuat file terpisah
// jika ingin menjalankan secara lokal dengan 'go run main.go'.
// Jika Anda hanya menggunakan Vercel CLI (vercel dev), Vercel akan menangani ini.

/*
// JANGAN sertakan fungsi main() di sini, karena ini adalah package 'handler'.
// Jika Anda ingin menjalankan lokal, buat file baru: backend/local/main.go
// dengan package main dan panggil handler.MainHandler
*/


// --- HANDLER FUNCTIONS ---

func searchCitiesHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query pencarian 'q' tidak boleh kosong"})
		return
	}

	apiURL := fmt.Sprintf("%s?q=%s&limit=5&appid=%s", geoAPIURL, query, openWeatherAPIKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Gagal terhubung ke layanan geocoding"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "Error dari layanan geocoding", "details": string(body)})
		return
	}

	var geoResults []geoResult
	if err := json.NewDecoder(resp.Body).Decode(&geoResults); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mem-parsing respons geocoding"})
		return
	}

	cleanedCities := make([]CleanedCity, 0)
	for _, result := range geoResults {
		parts := []string{result.Name}
		if result.State != "" && result.State != result.Name {
			parts = append(parts, result.State)
		}
		parts = append(parts, result.Country)

		cleanedCities = append(cleanedCities, CleanedCity{
			Name:        result.Name,
			DisplayName: strings.Join(parts, ", "),
			Lat:         result.Lat,
			Lon:         result.Lon,
			Country:     result.Country,
		})
	}

	c.JSON(http.StatusOK, cleanedCities)
}

func getWeatherHandler(c *gin.Context) {
	lat := c.Query("lat")
	lon := c.Query("lon")
	if lat == "" || lon == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'lat' dan 'lon' dibutuhkan"})
		return
	}

	apiURL := fmt.Sprintf("%s?lat=%s&lon=%s&appid=%s&units=metric&lang=id", forecastAPIURL, lat, lon, openWeatherAPIKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Gagal terhubung ke layanan cuaca"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "Error dari layanan cuaca", "details": string(body)})
		return
	}

	var weatherData WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mem-parsing data cuaca"})
		return
	}

	c.JSON(http.StatusOK, weatherData)
}

func getAirPollutionHandler(c *gin.Context) {
	lat := c.Query("lat")
	lon := c.Query("lon")
	if lat == "" || lon == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'lat' dan 'lon' dibutuhkan"})
		return
	}

	apiURL := fmt.Sprintf("%s?lat=%s&lon=%s&appid=%s", airPollutionAPIURL, lat, lon, openWeatherAPIKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Gagal terhubung ke layanan kualitas udara"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "Error dari layanan kualitas udara", "details": string(body)})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca respons dari layanan kualitas udara"})
		return
	}
	
	c.Data(http.StatusOK, "application/json", body)
}