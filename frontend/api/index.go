package handler // <-- PERUBAHAN DI SINI

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
var router *gin.Engine // Jadikan router sebagai variabel global

// Konstanta untuk URL API OpenWeatherMap
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
			Co  float64 `json:"co"`
			No2 float64 `json:"no2"`
			O3  float64 `json:"o3"`
			So2 float64 `json:"so2"`
		} `json:"components"`
	} `json:"list"`
}


// --- INISIALISASI ROUTER (Berjalan sekali saat fungsi di-load) ---

func init() {
	// Memuat environment variable (untuk development lokal)
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan. Menggunakan env sistem.")
	}
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	if openWeatherAPIKey == "" {
		log.Println("PERINGATAN: OPENWEATHER_API_KEY tidak diatur. Panggilan API akan gagal.")
	}

	// Inisialisasi Gin Router
	router = gin.Default()

	// Konfigurasi CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "OPTIONS"} 
	router.Use(cors.New(config))

	// Daftarkan rute langsung ke router utama
	router.GET("/search", searchCitiesHandler)
	router.GET("/weather", getWeatherHandler)
	router.GET("/air-pollution", getAirPollutionHandler)
	
	log.Println("Router Gin berhasil diinisialisasi untuk Vercel.")
}

// --- HANDLER UTAMA (Ini yang dipanggil Vercel) ---

func Handler(w http.ResponseWriter, r *http.Request) {
	// Teruskan permintaan ke router Gin yang sudah diinisialisasi
	router.ServeHTTP(w, r)
}

// --- FUNGSI UTAMA (HANYA UNTUK DEVELOPMENT LOKAL) ---
func main() {
	log.Println("Menjalankan server lokal di http://localhost:8080")
	// init() sudah berjalan, jadi kita tinggal jalankan router
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Gagal menjalankan server lokal: %v", err)
	}
}


// --- HANDLER & FUNGSI BANTUAN ---

func searchCitiesHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query pencarian 'q' tidak boleh kosong"})
		return
	}

	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak terkonfigurasi."})
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

	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak terkonfigurasi."})
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

	if openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API Key server tidak terkonfigurasi."})
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