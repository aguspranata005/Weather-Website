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
			Deg   float64 `json:"deg"` // <-- [PERBAIKAN] BARIS INI DITAMBAHKAN
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
// Struct ini dibuat agar cocok dengan apa yang diharapkan oleh frontend.
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

// --- HANDLER UTAMA UNTUK VERCEL ---
// Handler adalah fungsi utama yang akan dieksekusi oleh Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Setup router setiap kali ada request
	router, err := setupRouter()
	if err != nil {
		log.Printf("Error initializing router: %v", err)
		http.Error(w, "Gagal inisialisasi server", http.StatusInternalServerError)
		return
	}
	// Serahkan request ke Gin
	router.ServeHTTP(w, r)
}

// setupRouter berisi semua kode yang sebelumnya ada di func main()
func setupRouter() (*gin.Engine, error) {
	// Memuat .env. Ini hanya untuk lokal. Di Vercel, kita pakai Environment Variables.
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan.")
	}
	openWeatherAPIKey = os.Getenv("OPENWEATHER_API_KEY")
	if openWeatherAPIKey == "" {
		// Jangan pakai log.Fatal, kembalikan error
		return nil, fmt.Errorf("FATAL: Environment variable OPENWEATHER_API_KEY tidak diatur")
	}

	router := gin.Default()

	// Konfigurasi CORS
	config := cors.DefaultConfig()	
	config.AllowOrigins = []string{"*"} // Izinkan semua origin, Vercel akan menanganinya
	router.Use(cors.New(config))

	// BIARKAN GRUP /api INI SEPERTI ASLINYA
	// Vercel akan meneruskan path /api/... ke fungsi ini
	// api := router.Group("/api")
	// {
	// 	api.GET("/search", searchCitiesHandler)
	// 	api.GET("/weather", getWeatherHandler)
	// 	api.GET("/air-pollution", getAirPollutionHandler)
	// }

	router.GET("/search", searchCitiesHandler)
	router.GET("/weather", getWeatherHandler)
	router.GET("/air-pollution", getAirPollutionHandler)

	// Hapus router.Run(), cukup kembalikan router-nya
	return router, nil
}

// --- HANDLER & FUNGSI BANTUAN ---

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GGagal mem-parsing data cuaca"})
		return
	}

	c.JSON(http.StatusOK, weatherData)
}

// [DIPERBAIKI] getAirPollutionHandler sekarang mengirimkan respons yang sesuai dengan harapan frontend.
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
	// Ini memastikan struktur data (termasuk 'list') tetap sama
	c.Data(http.StatusOK, "application/json", body)
}