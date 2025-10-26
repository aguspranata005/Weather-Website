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

// Konstanta untuk URL API OpenWeatherMap agar mudah dikelola.
const (
	geoAPIURL          = "https://api.openweathermap.org/geo/1.0/direct"
	forecastAPIURL     = "https://api.openweathermap.org/data/2.5/forecast"
	airPollutionAPIURL = "https://api.openweathermap.org/data/2.5/air_pollution"
)

// --- STRUKTUR DATA (STRUCTS) ---
// (Struktur data Anda dari file asli tidak berubah)
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

// --- ENTRY POINT UNTUK VERCELL ---
// func main() diganti menjadi func Handler()
func Handler(w http.ResponseWriter, r *http.Request) {
	// Memuat environment variable dari file .env (hanya untuk lokal)
	// Di Vercel, variabel akan dimuat secara otomatis.
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan. Pastikan OPENWEATHER_API_KEY diatur di environment sistem.")
	}
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	if openWeatherAPIKey == "" {
		// Gunakan log.Println, JANGAN log.Fatal agar serverless tidak crash
		log.Println("FATAL: Environment variable OPENWEATHER_API_KEY tidak diatur.")
	}

	// Inisialisasi Gin Router.
	router := gin.Default()

	// Konfigurasi CORS (Cross-Origin Resource Sharing)
	config := cors.DefaultConfig()
	// *** PERUBAHAN PENTING: Izinkan semua origin untuk Vercel ***
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "OPTIONS"} // Izinkan metode yang diperlukan
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Grup routing untuk semua endpoint API di bawah prefix /api.
	// Ini cocok dengan rewrite Vercel: /api/...
	api := router.Group("/api")
	{
		api.GET("/search", searchCitiesHandler)
		api.GET("/weather", getWeatherHandler)
		api.GET("/air-pollution", getAirPollutionHandler)
	}

	// Menjalankan server
	// log.Println("Server backend berjalan di http://localhost:8080") // Tidak perlu di Vercel
	// router.Run(":8080") // HAPUS baris ini
	
	// *** TAMBAHKAN baris ini untuk Vercel ***
	// Menggunakan router Gin untuk melayani permintaan serverless
	router.ServeHTTP(w, r)
}

// --- HANDLER & FUNGSI BANTUAN ---

// searchCitiesHandler menangani permintaan pencarian kota.
func searchCitiesHandler(c *gin.Context) {
	// Tambahkan pengecekan API Key di dalam handler
	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak dikonfigurasi."})
		return
	}
	
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
	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak dikonfigurasi."})
		return
	}

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

// getAirPollutionHandler
func getAirPollutionHandler(c *gin.Context) {
	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak dikonfigurasi."})
		return
	}

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