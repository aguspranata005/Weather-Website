package main

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

// Konstanta untuk URL API OpenWeatherMap agar mudah dikelola.
const (
	geoAPIURL          = "https://api.openweathermap.org/geo/1.0/direct"
	forecastAPIURL     = "https://api.openweathermap.org/data/2.5/forecast"
	airPollutionAPIURL = "https://api.openweathermap.org/data/2.5/air_pollution"
)

// --- STRUKTUR DATA (STRUCTS) ---

// CleanedCity adalah struct untuk data kota yang sudah bersih dan siap dikirim ke frontend.
type CleanedCity struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Country     string  `json:"country"`
}

// geoResult adalah struct untuk menampung response mentah dari Geo API OpenWeatherMap.
type geoResult struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	State   string  `json:"state,omitempty"`
}

// WeatherResponse adalah struct lengkap untuk menampung response dari Forecast API.
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
		Pop float64 `json:"pop"` // Probability of precipitation
	} `json:"list"`
	City struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"city"`
}

// AirPollutionResponse adalah struct untuk menampung response dari Air Pollution API.
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

// --- FUNGSI UTAMA (MAIN) ---

func main() {
	// Memuat environment variable dari file .env (hanya untuk lokal).
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan. (Ini normal di produksi)")
	}
	
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	if openWeatherAPIKey == "" {
		log.Fatal("FATAL: Environment variable OPENWEATHER_API_KEY tidak diatur. Aplikasi tidak dapat berjalan.")
	}

	// Atur GIN_MODE ke "release" di produksi
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inisialisasi Gin Router.
	router := gin.Default()

	// --- Konfigurasi CORS (PENTING UNTUK DEPLOYMENT) ---
	// Ambil URL frontend dari environment variable
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Fallback untuk lokal
	}

	config := cors.DefaultConfig()
	// Izinkan localhost DAN URL frontend produksi
	config.AllowOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173", frontendURL}
	router.Use(cors.New(config))
	
	// Izinkan semua proxy (diperlukan untuk Railway)
	router.SetTrustedProxies(nil)

	// Grup routing untuk semua endpoint API di bawah prefix /api.
	api := router.Group("/api")
	{
		api.GET("/search", searchCitiesHandler)
		api.GET("/weather", getWeatherHandler)
		api.GET("/air-pollution", getAirPollutionHandler)
	}

	// --- Konfigurasi Port (PENTING UNTUK DEPLOYMENT) ---
	// Ambil PORT dari environment variable yang disediakan Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback jika PORT tidak diatur (untuk lokal)
	}

	// Menjalankan server
	log.Printf("Server backend berjalan di port :%s", port)
	router.Run(":" + port) // Gunakan port dari Railway
}

// --- HANDLER & FUNGSI BANTUAN ---

// (Semua fungsi handler Anda yang lain tetap sama)

// searchCitiesHandler menangani permintaan pencarian kota.
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

// getWeatherHandler menangani permintaan data cuaca lengkap.
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

// getAirPollutionHandler menangani permintaan data polusi udara.
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

	// Membaca body respons untuk diteruskan langsung
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca respons dari layanan kualitas udara"})
		return
	}

	// Mengirimkan kembali body JSON mentah dari OpenWeatherMap
	c.Data(http.StatusOK, "application/json", body)
}