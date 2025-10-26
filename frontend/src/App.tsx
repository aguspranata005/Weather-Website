import { useState, useEffect, useCallback, useRef } from 'react';
import './App.css'; // Pastikan impor CSS ini ada

// --- PERBAIKAN: Impor ThemeContext dari file eksternal ---
// Hapus semua definisi 'ThemeContextType', 'ThemeContext', 'ThemeProvider', 'useTheme' 
// yang ada di dalam file App.tsx sebelumnya (baris 6-63).
import { useTheme } from './ThemeContext.tsx'; // <-- TAMBAHKAN IMPORT INI

// --- Definisi Tipe ---
// (Definisi tipe Anda tidak berubah)
interface CleanedCity {
  lat: number;
  lon: number;
  displayName: string;
}

interface WeatherListItem {
  dt: number;
  main: {
    temp: number;
    feels_like: number;
    humidity: number;
  };
  weather: {
    description: string;
    icon: string;
  }[];
  wind: {
    speed: number;
    deg: number;
  };
  pop: number;
}

interface WeatherResponse {
  list: WeatherListItem[];
  city: {
    name: string;
    country: string;
    sunrise: number;
    sunset: number;
  };
}

interface AirPollutionResponse {
  list: {
    main: {
      aqi: number;
    };
    components: {
      co: number;
      no2: number;
      o3: number;
      so2: number;
    };
  }[];
}


// --- Konfigurasi & Fungsi Helper ---
// (Semua fungsi helper Anda tidak berubah)
const API_BASE_URL = '/api'; 
const getWeatherIconUrl = (iconCode: string): string => {
  const baseUrl = "https://www.amcharts.com/wp-content/themes/amcharts4/css/img/icons/weather/animated/";
  let iconName: string;

  switch (iconCode) {
    case '01d': iconName = 'day.svg'; break;
    case '01n': iconName = 'night.svg'; break;
    case '02d': iconName = 'cloudy-day-1.svg'; break;
    case '02n': iconName = 'cloudy-night-1.svg'; break;
    case '03d': case '03n': iconName = 'cloudy.svg'; break;
    case '04d': case '04n': iconName = 'cloudy.svg'; break;
    case '09d': case '09n': iconName = 'rainy-1.svg'; break;
    case '10d': case '10n': iconName = 'rainy-6.svg'; break;
    case '11d': case '11n': iconName = 'thunder.svg'; break;
    case '13d': case '13n': iconName = 'snowy-6.svg'; break;
    case '50d': case '50n': iconName = 'fog.svg'; break;
    default: iconName = 'weather.svg'; break; 
  }
  return `${baseUrl}${iconName}`;
};

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

const getAqiClassName = (aqi: number): string => {
  switch (aqi) {
    case 1: return 'aqi-baik';
    case 2: return 'aqi-cukup';
    case 3: return 'aqi-sedang';
    case 4: return 'aqi-buruk';
    case 5: return 'aqi-sangat-buruk';
    default: return 'aqi-unknown';
  }
};

const getWindDirectionText = (deg: number): string => {
  const directions = [
    'U', 'UTL', 'TL', 'TTL',
    'T', 'TG', 'STG', 'S',
    'SBD', 'BD', 'BBD', 'B',
    'UBL', 'BL', 'UBL', 'U'
  ];
  
  const index = Math.round(((deg % 360) / 360) * 16) % 16;
  return directions[index] || 'Utara';
};  


// --- Komponen Inti Aplikasi (agar bisa menggunakan useTheme) ---
function AppContent() {
  // useTheme() sekarang diimpor dari file ThemeContext.tsx
  const { theme, toggleTheme } = useTheme(); 
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<CleanedCity[]>([]);
  const [weatherData, setWeatherData] = useState<WeatherResponse | null>(null);
  const [airQualityData, setAirQualityData] = useState<AirPollutionResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [locationLoading, setLocationLoading] = useState(false);
  const [isSearchActive, setIsSearchActive] = useState(false); 
  const searchInputRef = useRef<HTMLInputElement>(null);

  // --- Fungsi Fetch Data ---
  const fetchAllData = useCallback(async (lat: number, lon: number) => {
    setLoading(true);
    setError(null);
    setWeatherData(null);
    setAirQualityData(null);
    setSearchResults([]); 

    const safeJsonParse = async (response: Response) => {
      const contentType = response.headers.get("content-type");
      if (response.ok && contentType && contentType.includes("application/json")) {
        return response.json();
      }
      const errorText = await response.text();
      throw new Error(`Respons server tidak valid. Status: ${response.status}. Pesan: ${errorText.substring(0, 200)}...`);
    };

    try {
      const [weatherResponse, airPollutionResponse] = await Promise.all([
        fetch(`${API_BASE_URL}/weather?lat=${lat}&lon=${lon}`),
        fetch(`${API_BASE_URL}/air-pollution?lat=${lat}&lon=${lon}`)
      ]);

      const weatherDataResult = await safeJsonParse(weatherResponse);
      const airDataResult = await safeJsonParse(airPollutionResponse);

      setWeatherData(weatherDataResult);
      setAirQualityData(airDataResult);

    } catch (err: any) {
      console.error("Gagal mengambil data:", err); 
      setError(err.message || "Terjadi kesalahan saat mengambil data."); 
    } finally {
      setLoading(false);
      setLocationLoading(false); 
    }
  }, []); 

  // --- Fungsi Cari Kota ---
  const fetchCities = async (query: string) => {
    if (query.length < 3) {
      setSearchResults([]); 
      return;
    }
    try {
      const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`);
      if (!response.ok) {
        console.error("Error mencari kota:", response.statusText);
        setSearchResults([]);
        return;
      }
      const data: CleanedCity[] = await response.json();
      setSearchResults(data);
    } catch (error) {
      console.error("Gagal mencari kota:", error);
      setSearchResults([]); 
    }
  };

  // --- Handler Event ---
  const handleCitySelect = (city: CleanedCity) => {
    setSearchQuery(city.displayName); 
    setSearchResults([]); 
    fetchAllData(city.lat, city.lon); 
    setIsSearchActive(false); 
  };

  const handleUseMyLocation = useCallback(() => {
    if (navigator.geolocation) {
      setLocationLoading(true);
      setError(null); 
      setSearchQuery(''); 
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const { latitude, longitude } = position.coords;
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
      setError("Geolocation tidak didukung oleh browser ini.");
      setLoading(false); 
    }
  }, [fetchAllData]); 

  const handleClearSearch = () => {
    setSearchQuery('');
    setSearchResults([]);
  };

  // --- Effect untuk Fetch Data Awal ---
  useEffect(() => {
    handleUseMyLocation(); 
  }, [handleUseMyLocation]); 

  // --- Data yang akan Dirender ---
  const todayWeather = weatherData?.list?.[0];
  const todayAirQuality = airQualityData?.list?.[0];
  const todayDate = new Date().toLocaleDateString('id-ID', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric'
  });

  // --- Render ---
  // (Seluruh bagian 'return' dari AppContent tetap SAMA PERSIS seperti file asli Anda)
  // ... (Salin semua kode JSX dari baris 251 hingga 624 di file App.tsx asli Anda)
  return (
    <div className="container">
      <header className={`header ${isSearchActive ? 'search-active' : ''}`}>
        {/* Tombol Search/Back Mobile */}
        <div 
          className="mobile-search-toggle"
          onClick={() => {
            if (isSearchActive) {
              setIsSearchActive(false);
              handleClearSearch();
              searchInputRef.current?.blur();
            } else {
              setIsSearchActive(true);
              setTimeout(() => {
                searchInputRef.current?.focus();
              }, 50);
            }
          }}
        >
          <span className="material-symbols-rounded mobile-search-icon">search</span>
          <span className="material-symbols-rounded mobile-back-icon">arrow_back</span>
        </div>

        {/* Search Bar */}
        <div className="search-container">
          <span className="material-symbols-rounded search-icon-left">search</span>
          <input
            ref={searchInputRef}
            type="text"
            className="search-input"
            placeholder="Cari Tempat..."
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              fetchCities(e.target.value);
            }}
          />
          {searchQuery.length > 0 && (
            <span className="material-symbols-rounded search-icon-right" onClick={handleClearSearch}>close</span>
          )}
          {searchResults.length > 0 && (
            <ul className="search-results">
              {searchResults.map((city) => (
                <li key={`${city.lat}-${city.lon}`} onClick={() => handleCitySelect(city)}>
                  {city.displayName}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Tombol Aksi */}
        <div className="header-actions">
          <button className="location-button" onClick={handleUseMyLocation} disabled={locationLoading || loading} aria-label="Gunakan Lokasi Saya">
            {locationLoading ? <div className="loading-spinner"></div> : <span className="material-symbols-rounded">my_location</span>}
          </button>
          <button className="location-button theme-toggle-button" onClick={toggleTheme} aria-label="Ganti tema">
            <span className="material-symbols-rounded">{theme === 'light' ? 'dark_mode' : 'light_mode'}</span>
          </button>
        </div>
      </header>

      <main className="dashboard">
        {loading && !error && ( 
          <div className="loading-container">
            <div className="loading-spinner large"></div>
            <p>Mengambil data terbaru...</p>
          </div>
        )}
        {error && ( 
          <div className="error-message">
            <span className="material-symbols-rounded">sentiment_dissatisfied</span> 
            <strong>Gagal Memuat Data</strong>
            <p>{error}</p>
            {error !== "Geolocation tidak didukung oleh browser ini." && (
                 <button className="retry-button" onClick={handleUseMyLocation}>
                    Coba Lagi dengan Lokasi Saya
                 </button>
            )}
            <p style={{marginTop: '4px', fontSize: '0.9em', color: 'var(--on-surface-variant)'}}>
                Atau gunakan kolom pencarian di bagian atas layar.
            </p>
          </div>
        )}
        {weatherData && todayWeather && todayAirQuality && !error && !loading && ( 
          <>
            {/* Kolom Kiri */}
            <section className="main-content">
              {/* Kartu Cuaca Saat Ini */}
              <div className="current-weather">
                <h2>Cuaca Saat Ini</h2>
                <div className="current-temp-container">
                  <div className="temp-desc-wrapper">
                    <p className="current-temp">{Math.round(todayWeather.main.temp)}°C</p>
                  </div>
                  <img src={getWeatherIconUrl(todayWeather.weather[0].icon)} alt={todayWeather.weather[0].description} className="weather-icon" />
              </div>
                 <div className="description-of-weather"> 
                    <p className="current-desc">{todayWeather.weather[0].description}</p>
                 </div>
                <div className="location-and-date">
                  <div className="info-item">
                    <span className="material-symbols-rounded">location_on</span>
                    <span>{weatherData.city.name}, {weatherData.city.country}</span>
                  </div>
                  <div className="info-item">
                    <span className="material-symbols-rounded">calendar_today</span>
                    <span>{todayDate}</span>
                  </div>
                </div>
              </div>

              {/* Kartu Ramalan 24 Jam */}
              <div className="forecast-card">
                <div className="aq-title-wrapper">
                  <span className="material-symbols-rounded aq-icon">avg_pace</span>
                  <h2>Ramalan 24 Jam</h2>
                </div>
                <div className="hourly-forecast">
                  {weatherData.list.slice(0, 8).map(item => (
                    <HourCard key={item.dt} item={item} />
                  ))}
                </div>
                <div className="hourly-wind-forecast">
                  {weatherData.list.slice(0, 8).map(item => (
                    <WindHourCard key={`wind-${item.dt}`} item={item} />
                  ))}
                </div>
              </div>
            </section>

            {/* Kolom Kanan */}
            <aside className="sidebar">
              {/* Kartu Sorotan */}
              <div className="highlights">
                 <div className="aq-title-wrapper"> 
                    <span className="material-symbols-rounded highlight-title-icon">trending_up</span>
                    <h2>Sorotan Hari Ini</h2>
                 </div>
   B            <div className="highlight-grid">
                  <HighlightCard icon="humidity_percentage" title="Kelembapan" value={`${todayWeather.main.humidity}%`} />
                  <HighlightCard icon="thermostat" title="Terasa Seperti" value={`${Math.round(todayWeather.main.feels_like)}°C`} />
                  <HighlightCard icon="umbrella" title="Peluang Hujan" value={`${Math.round(todayWeather.pop * 100)}%`} />
                  <HighlightCard icon="sunny" title="Matahari Terbit" value={new Date(weatherData.city.sunrise * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} />
                  <HighlightCard icon="wb_twilight" title="Matahari Terbenam" value={new Date(weatherData.city.sunset * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} />
                </div>
              </div>

               {/* Kartu Kualitas Udara */}
               <div className="forecast-card"> 
                 <div className="aq-title-wrapper">
                     <span className="material-symbols-rounded aq-icon">airwave</span>
                     <h2>Indeks Kualitas Udara</h2>
                 </div>
                 <AirQualityCard aqi={todayAirQuality.main.aqi} components={todayAirQuality.components} />
               </div>


              {/* Kartu Ramalan 5 Hari */}
                <div className="forecast-card">
                  <div className="aq-title-wrapper">
                    <span className="material-symbols-rounded aq-icon">cloud</span>
                    <h2>Ramalan 5 Hari</h2>
                  </div>
                  
                  <div className="daily-forecast">
                    {process5DayForecast(weatherData.list).map(day => (
                      <DayForecastItem key={day.date} day={day} />
                    ))}
                  </div>
                </div>
            </aside>
          </>
        )}
      </main>
    </div>
  );
}

// --- Sub-Komponen (Helper Components) ---
// (Semua sub-komponen Anda: HighlightCard, AirQualityCard, HourCard, WindHourCard,
// process5DayForecast, dan DayForecastItem tetap SAMA PERSIS seperti file asli Anda)
// ... (Salin semua kode dari baris 627 hingga 752 di file App.tsx asli Anda)
const HighlightCard = ({ icon, title, value }: { icon: string; title: string; value: string }) => (
  <div className="highlight-item">
    <span className="material-symbols-rounded">{icon}</span>
    <div>
      <p className="highlight-title">{title}</p>
      <p className="highlight-value">{value}</p>
    </div>
  </div>
);

const AirQualityCard = ({ aqi, components }: { aqi: number, components: AirPollutionResponse['list'][0]['components'] }) => (
  <div className="air-quality-card">
    <div className="aq-summary">
        <p className={`aq-value ${getAqiClassName(aqi)}`}>
            {getAqiDescription(aqi)} 
        </p>
        </div>
    <div className="aq-components-grid">
      <div className="aq-component"><span>CO</span><p>{components?.co?.toFixed(1) ?? '-'} <small>μg/m³</small></p></div>
      <div className="aq-component"><span>NO₂</span><p>{components?.no2?.toFixed(1) ?? '-'} <small>μg/m³</small></p></div>
      <div className="aq-component"><span>O₃</span><p>{components?.o3?.toFixed(1) ?? '-'} <small>μg/m³</small></p></div>
      <div className="aq-component"><span>SO₂</span><p>{components?.so2?.toFixed(1) ?? '-'} <small>μg/m³</small></p></div>
    </div>
  </div>
);


const HourCard = ({ item }: { item: WeatherListItem }) => (
  <div className="hour-item">
    <p className="hour-time">
      {new Date(item.dt * 1000).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
    </p>
    <img src={getWeatherIconUrl(item.weather?.[0]?.icon || '')} alt={item.weather?.[0]?.description || 'Cuaca'} width="40" height="40"/>
    <p className="hour-temp">{Math.round(item.main.temp)}°</p> 
  </div>
);

const WindHourCard = ({ item }: { item: WeatherListItem }) => (
  <div className="wind-hour-item">
    <span className="wind-direction-text">{getWindDirectionText(item.wind.deg)}</span>
    <span
      className="material-symbols-rounded wind-direction-icon"
      style={{ transform: `rotate(${item.wind.deg || 0}deg)` }}
    >
      navigation
    </span>
    <span className="wind-speed">{item.wind.speed?.toFixed(1) ?? '-'} m/s</span>
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
  if (!list || list.length === 0) return []; 

  const dailyData: { [key: string]: { temps: number[], icons: string[] } } = {};

  list.forEach(item => {
     if (item && item.dt) { 
        const dateStr = new Date(item.dt * 1000).toISOString().split('T')[0];
        if (!dailyData[dateStr]) {
            dailyData[dateStr] = { temps: [], icons: [] };
        }
        if (item.main && typeof item.main.temp === 'number') {
           dailyData[dateStr].temps.push(item.main.temp);
        }
        if (item.weather?.[0]?.icon) {
           dailyData[dateStr].icons.push(item.weather[0].icon);
        }
     }
  });


  return Object.keys(dailyData)
    .filter(date => dailyData[date].temps.length > 0 && dailyData[date].icons.length > 0) 
    .slice(0, 5) 
    .map(date => {
      const dayTemps = dailyData[date].temps;
      const dayIcons = dailyData[date].icons;

      const iconFrequency = dayIcons.reduce((acc, icon) => {
          acc[icon] = (acc[icon] || 0) + 1;
          return acc;
        }, {} as { [key: string]: number });
      const representativeIcon = Object.keys(iconFrequency).reduce(
          (a, b) => (iconFrequency[a] > iconFrequency[b] ? a : b), dayIcons[0] || ''); 

       const dayName = new Date(date + 'T00:00:00Z').toLocaleDateString('id-ID', { weekday: 'short', timeZone: 'UTC' }); 

      return {
        date: date,
        dayName: dayName,
        temp_max: Math.round(Math.max(...dayTemps)),
        temp_min: Math.round(Math.min(...dayTemps)),
        icon: representativeIcon.replace('n', 'd'),
      };
    });
};


const DayForecastItem = ({ day }: { day: ProcessedDay }) => (
  <div className="day-item">
    <img src={getWeatherIconUrl(day.icon || '')} alt="Ikon cuaca" width="40" height="40" />
    <span className="day-name">{day.dayName}</span>
    <span className="day-temp">{day.temp_max}° / {day.temp_min}°</span>
  </div>
);

// --- Komponen App utama yang membungkus dengan Provider ---

// --- PERBAIKAN: Hapus <ThemeProvider> dari sini ---
// Komponen App sekarang hanya me-render AppContent.
// Pembungkus <ThemeProvider> sudah ada di main.tsx
const App = () => (
  <AppContent />
);

export default App;