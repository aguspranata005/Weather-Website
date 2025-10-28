// Pastikan ini adalah `package main`
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http" // <-- Diperlukan untuk Vercel Handler
	"os"
	"strings"
	"sync" // <-- Diperlukan untuk inisialisasi router

	// "github.com/gin-contrib/cors" // <-- HAPUS: Tidak perlu lagi
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// --- KONFIGURASI & VARIABEL GLOBAL ---

var openWeatherAPIKey string
var (
	router *gin.Engine
	once   sync.Once // Variabel untuk memastikan router hanya di-setup sekali
)

const (
	geoAPIURL          = "https://api.openweathermap.org/geo/1.0/direct"
	forecastAPIURL     = "https://api.openweathermap.org/data/2.5/forecast"
	airPollutionAPIURL = "https://api.openweathermap.org/data/2.5/air_pollution"
)

// --- STRUCTS ---
// (Semua struct Anda dari file asli tetap sama)
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
			Co  float64 `json:"co"`
			No2 float64 `json:"no2"`
			O3  float64 `json:"o3"`
			So2 float64 `json:"so2"`
		} `json:"components"`
	} `json:"list"`
}
// --- AKHIR STRUCTS ---

// setupRouter berisi semua logika inisialisasi dari 'main()' lama Anda
func setupRouter() *gin.Engine {
	// Memuat .env (berguna untuk dev lokal, Vercel akan mengabaikannya)
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan (ini normal di Vercel).")
	}
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	if openWeatherAPIKey == "" {
		log.Fatal("FATAL: Environment variable OPENWEATHER_API_KEY tidak diatur.")
	}

	r := gin.Default()

	// --- HAPUS KONFIGURASI CORS ---
	// 'rewrites' di vercel.json membuat ini tidak perlu lagi.
	// config := cors.DefaultConfig()
	// config.AllowOrigins = []string{"...localhost..."}
	// r.Use(cors.New(config))

	// Prefix /api harus ada di sini, karena 'rewrites' Vercel meneruskannya
	api := r.Group("/api")
	{
		api.GET("/search", searchCitiesHandler)
		api.GET("/weather", getWeatherHandler)
		api.GET("/air-pollution", getAirPollutionHandler)
	}

	return r
}

// !! PENTING: Entrypoint untuk Vercel !!
// 'func main()' diganti dengan 'func Handler'
func Handler(w http.ResponseWriter, r *http.Request) {
	// Inisialisasi router hanya sekali
	once.Do(func() {
		router = setupRouter()
	})
	// Serahkan semua permintaan ke router Gin
	router.ServeHTTP(w, r)
}

// --- FUNGSI HANDLER (TETAP SAMA) ---
// Semua fungsi handler Anda (searchCitiesHandler, getWeatherHandler,
// getAirPollutionHandler) tetap sama persis seperti di file asli Anda.

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GGagal mem-parsing data cuaca"})
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