import { useState, useEffect, useCallback } from 'react';
import './App.css';

// Menggunakan 'import type' untuk mengimpor tipe data saja.
import type { CleanedCity, WeatherResponse, WeatherListItem } from './types';

// --- Tipe Data Baru untuk Kualitas Udara ---
export interface AirPollutionResponse {
  list: {
    main: {
      aqi: number; // Indeks Kualitas Udara (skala 1-5)
    };
    components: {
      co: number;   // Karbon monoksida (μg/m³)
      no2: number;  // Nitrogen dioksida (μg/m³)
      o3: number;   // Ozon (μg/m³)
      so2: number;  // Sulfur dioksida (μg/m³)
    };
  }[];
}

// --- Konfigurasi & Fungsi Helper ---
const API_BASE_URL = 'http://localhost:8080/api';
const getWeatherIconUrl = (iconCode: string) => `https://openweathermap.org/img/wn/${iconCode}@4x.png`;

const getAqiDescription = (aqi: number): string => {
  switch (aqi) {
    case 1: return 'Baik';
    case 2: return 'Cukup';
    case 3: return 'Sedang';
    case 4: return 'Buruk';
    case 5: return 'Sangat Buruk';
    default: return 'Tidak Diketahui';
  }
};

const degreesToCardinal = (deg: number): string => {
  // Standar 16 arah mata angin internasional
  const directions = [
    'N', 'NNE', 'NE', 'ENE',
    'E', 'ESE', 'SE', 'SSE',
    'S', 'SSW', 'SW', 'WSW',
    'W', 'WNW', 'NW', 'NNW'
  ];
  const index = Math.floor((deg + 11.25) / 22.5) % 16;
  return directions[index];
};


// --- Komponen Utama Aplikasi ---
function App() {
  // --- State Management ---
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<CleanedCity[]>([]);
  const [weatherData, setWeatherData] = useState<WeatherResponse | null>(null);
  const [airQualityData, setAirQualityData] = useState<AirPollutionResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [locationLoading, setLocationLoading] = useState(false);

  // --- Fungsi untuk Mengambil Data ---
  const fetchAllData = useCallback(async (lat: number, lon: number) => {
    setLoading(true);
    setError(null);
    setWeatherData(null);
    setAirQualityData(null);
    setSearchResults([]);

    try {
      const [weatherResponse, airPollutionResponse] = await Promise.all([
        fetch(`${API_BASE_URL}/weather?lat=${lat}&lon=${lon}`),
        fetch(`${API_BASE_URL}/air-pollution?lat=${lat}&lon=${lon}`)
      ]);

      if (!weatherResponse.ok) throw new Error((await weatherResponse.json()).details || 'Gagal mengambil data cuaca.');
      if (!airPollutionResponse.ok) throw new Error((await airPollutionResponse.json()).details || 'Gagal mengambil data kualitas udara.');

      const weatherDataResult: WeatherResponse = await weatherResponse.json();
      const airDataResult: AirPollutionResponse = await airPollutionResponse.json();

      setWeatherData(weatherDataResult);
      setAirQualityData(airDataResult);

    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
      setLocationLoading(false);
    }
  }, []);

  // --- Fungsi untuk Mencari Kota ---
  const fetchCities = async (query: string) => {
    if (query.length < 3) {
      setSearchResults([]);
      return;
    }
    try {
      const response = await fetch(`${API_BASE_URL}/search?q=${query}`);
      if (!response.ok) return;
      const data: CleanedCity[] = await response.json();
      setSearchResults(data);
    } catch (error) {
      console.error("Gagal mencari kota:", error);
    }
  };

  const handleCitySelect = (city: CleanedCity) => {
    setSearchQuery(city.displayName);
    fetchAllData(city.lat, city.lon);
  };

  const handleUseMyLocation = useCallback(() => {
    if (navigator.geolocation) {
      setLocationLoading(true);
      setError(null);
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const { latitude, longitude } = position.coords;
          setSearchQuery("Lokasi Saat Ini");
          fetchAllData(latitude, longitude);
        },
        (err) => {
          setLoading(false);
          setLocationLoading(false);
          setWeatherData(null);
          setAirQualityData(null);
          let errorMessage = "Tidak bisa mengakses lokasi.";
          switch (err.code) {
            case err.PERMISSION_DENIED:
              errorMessage = "Akses lokasi ditolak. Mohon izinkan akses lokasi di browser Anda.";
              break;
            case err.POSITION_UNAVAILABLE:
              errorMessage = "Informasi lokasi tidak tersedia.";
              break;
            case err.TIMEOUT:
              errorMessage = "Request lokasi timeout. Silakan coba lagi.";
              break;
          }
          setError(errorMessage);
        },
        { enableHighAccuracy: true, timeout: 10000, maximumAge: 300000 }
      );
    } else {
      setLoading(false);
      setError("Geolocation tidak didukung oleh browser ini.");
    }
  }, [fetchAllData]);

  useEffect(() => {
    handleUseMyLocation();
  }, [handleUseMyLocation]);

  const todayWeather = weatherData?.list[0];
  const todayAirQuality = airQualityData?.list[0];
  const todayDate = new Date().toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  });

  return (
    <div className="container">
      <header className="header">
        <div className="search-container">
          <md-outlined-text-field
            label="Cari Kota..."
            value={searchQuery}
            onInput={(e: any) => {
              setSearchQuery(e.target.value);
              fetchCities(e.target.value);
            }}
          />
          {searchResults.length > 0 && (
            <md-list className="search-results">
              {searchResults.map((city) => (
                <md-list-item
                  key={`${city.lat}-${city.lon}`}
                  headline={city.displayName}
                  onClick={() => handleCitySelect(city)}
                />
              ))}
            </md-list>
          )}
        </div>
        <button
          className="location-button"
          onClick={handleUseMyLocation}
          disabled={locationLoading || loading}
        >
          {locationLoading ? (
            <>
              <div className="loading-spinner"></div>
              <span>Mencari...</span>
            </>
          ) : (
            <>
              <span className="material-symbols-outlined">my_location</span>
              <span>Lokasi Saya</span>
            </>
          )}
        </button>
      </header>

      <main className="dashboard">
        {loading && (
          <div className="loading-container">
            <div className="loading-spinner large"></div>
            <p>Mengambil data terbaru...</p>
          </div>
        )}
        {error && (
          <div className="error-message">
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
              <span className="material-symbols-outlined">error</span>
              <strong>Terjadi Kesalahan</strong>
            </div>
            <p>{error}</p>
            <button className="retry-button" onClick={handleUseMyLocation}>
              Coba Lagi
            </button>
          </div>
        )}
        {weatherData && todayWeather && todayAirQuality && !error && (
          <>
            {/* Layout Grid: Kiri dan Kanan */}
            <div className="main-grid">
              <div className="left-column">
                {/* Kartu Cuaca Saat Ini */}
                <div className="current-weather">
                  <h2>Cuaca Saat Ini</h2>
                  <div className="current-temp-container">
                    <div className="temp-desc-wrapper">
                      <p className="current-temp">{Math.round(todayWeather.main.temp)}°C</p>
                    </div>
                    <img
                      src={getWeatherIconUrl(todayWeather.weather[0].icon)}
                      alt={todayWeather.weather[0].description}
                      className="weather-icon"
                    />
                  </div>
                  <div className="description-of-weather">
                    <p className="current-desc">{todayWeather.weather[0].description}</p>
                  </div>
                  <div className="location-and-date">
                    <div className="info-item">
                      <span className="material-symbols-outlined">location_on</span>
                      <span>{weatherData.city.name}, {weatherData.city.country}</span>
                    </div>
                    <div className="info-item">
                      <span className="material-symbols-outlined">calendar_today</span>
                      <span>{todayDate}</span>
                    </div>
                  </div>
                </div>

                {/* Kartu Ramalan 5 Hari */}
                <div className="forecast-card">
                  <h2>Ramalan 5 Hari</h2>
                  <md-list>
                    {process5DayForecast(weatherData.list).map(day => (
                      <DayForecastItem key={day.date} day={day} />
                    ))}
                  </md-list>
                </div>
              </div>

              <div className="right-column">
                {/* Kartu Sorotan Hari Ini */}
                <div className="highlights">
                  <h2>Sorotan Hari Ini</h2>
                  <div className="highlight-grid">
                    <HighlightCard icon="air" title="Kecepatan Angin" value={`${todayWeather.wind.speed.toFixed(1)} m/s`} />
                    <HighlightCard icon="humidity_percentage" title="Kelembapan" value={`${todayWeather.main.humidity}%`} />
                    <HighlightCard icon="thermostat" title="Terasa Seperti" value={`${Math.round(todayWeather.main.feels_like)}°C`} />
                    <HighlightCard icon="umbrella" title="Peluang Hujan" value={`${Math.round(todayWeather.pop * 100)}%`} />
                    <HighlightCard icon="sunny" title="Matahari Terbit" value={new Date(weatherData.city.sunrise * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} />
                    <HighlightCard icon="wb_twilight" title="Matahari Terbenam" value={new Date(weatherData.city.sunset * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} />
                  </div>
                  <hr className="divider" />
                  <AirQualityCard aqi={todayAirQuality.main.aqi} components={todayAirQuality.components} />
                </div>

                {/* Kartu Ramalan 24 Jam */}
                <div className="forecast-card">
                  <h2>Ramalan 24 Jam</h2>
                  <div className="hourly-forecast">
                    {weatherData.list.slice(0, 8).map(item => (
                      <HourCard key={item.dt} item={item} />
                    ))}
                  </div>
                  <div className="hourly-forecast">
                    {weatherData.list.slice(0, 8).map(item => (
                      <HourWindCard key={item.dt} item={item} />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </>
        )}
      </main>
    </div>
  );
}

// --- Sub-Komponen ---
const HighlightCard = ({ icon, title, value }: { icon: string; title: string; value: string }) => (
  <div className="highlight-item">
    <span className="material-symbols-outlined">{icon}</span>
    <div>
      <p className="highlight-title">{title}</p>
      <p className="highlight-value">{value}</p>
    </div>
  </div>
);

const AirQualityCard = ({ aqi, components }: { aqi: number, components: AirPollutionResponse['list'][0]['components'] }) => (
  <div className="air-quality-card">
    <h3 className="aq-title">Indeks Kualitas Udara</h3>
    <div className="aq-summary">
      <span className="material-symbols-outlined aq-icon">airwave</span>
      <p className="aq-value">{getAqiDescription(aqi)}</p>
    </div>
    <div className="aq-components-grid">
      <div className="aq-component">
        <span>CO</span>
        <p>{components.co.toFixed(1)} <span>μg/m³</span></p>
      </div>
      <div className="aq-component">
        <span>NO₂</span>
        <p>{components.no2.toFixed(1)} <span>μg/m³</span></p>
      </div>
      <div className="aq-component">
        <span>O₃</span>
        <p>{components.o3.toFixed(1)} <span>μg/m³</span></p>
      </div>
      <div className="aq-component">
        <span>SO₂</span>
        <p>{components.so2.toFixed(1)} <span>μg/m³</span></p>
      </div>
    </div>
  </div>
);

const HourCard = ({ item }: { item: WeatherListItem }) => (
  <div className="hour-item">
    <p>{new Date(item.dt * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</p>
    <img src={getWeatherIconUrl(item.weather[0].icon)} alt={item.weather[0].description} width="50" />
    <p className="hour-temp">{Math.round(item.main.temp)}°C</p>
  </div>
);

const HourWindCard = ({ item }: { item: WeatherListItem }) => (
  <div className="hour-wind-item">
    <p className="hour-time">{new Date(item.dt * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</p>
     <div className="wind-direction-icon-wrapper">
      {/* IKON PANAH YANG BERPUTAR */}
      <span
        className="material-symbols-outlined wind-direction-icon"
        style={{ transform: `rotate(${item.wind.deg}deg)` }}
      >
        navigation
      </span>
    </div>
    <p className="hour-wind-speed">{item.wind.speed.toFixed(1)} m/s</p>
    <p className="hour-wind-direction">{degreesToCardinal(item.wind.deg)}</p>
  </div>
);

interface ProcessedDay {
  date: string;
  dayName: string;
  temp_max: number;
  temp_min: number;
  icon: string;
}

const process5DayForecast = (list: WeatherListItem[]): ProcessedDay[] => {
  const dailyData: { [key: string]: { temps: number[], icons: string[] } } = {};
  list.forEach(item => {
    const date = new Date(item.dt * 1000).toISOString().split('T')[0];
    if (!dailyData[date]) {
      dailyData[date] = { temps: [], icons: [] };
    }
    dailyData[date].temps.push(item.main.temp);
    dailyData[date].icons.push(item.weather[0].icon);
  });

  return Object.keys(dailyData).slice(0, 5).map(date => {
    const dayTemps = dailyData[date].temps;
    const dayIcons = dailyData[date].icons;
    const iconFrequency = dayIcons.reduce((acc, icon) => ({ ...acc, [icon]: (acc[icon] || 0) + 1 }), {} as { [key: string]: number });
    const representativeIcon = Object.keys(iconFrequency).reduce((a, b) => iconFrequency[a] > iconFrequency[b] ? a : b);

    return {
      date: date,
      dayName: new Date(date).toLocaleDateString('id-ID', { weekday: 'long' }),
      temp_max: Math.round(Math.max(...dayTemps)),
      temp_min: Math.round(Math.min(...dayTemps)),
      icon: representativeIcon.replace('n', 'd'),
    };
  });
}

const DayForecastItem = ({ day }: { day: ProcessedDay }) => (
  <>
    <div className="day-item">
      <img src={getWeatherIconUrl(day.icon)} alt="" width="50" />
      <span className="day-name">{day.dayName}</span>
      <span className="day-temp">{day.temp_max}° / {day.temp_min}°</span>
    </div>
    <md-divider />
  </>
);

export default App;