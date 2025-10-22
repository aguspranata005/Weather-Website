// src/types.ts

// Tipe untuk data kota yang sudah dibersihkan dari API Geocoding
export interface CleanedCity {
  lat: number;
  lon: number;
  displayName: string;
}

// Tipe untuk setiap item dalam daftar ramalan cuaca per 3 jam
export interface WeatherListItem {
  dt: number; // Timestamp
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
  pop: number; // Probability of precipitation
}

// Tipe untuk respons lengkap dari API cuaca
export interface WeatherResponse {
  list: WeatherListItem[];
  city: {
    name: string;
    country: string;
    sunrise: number; // Timestamp
    sunset: number;  // Timestamp
  };
}

// Tipe untuk respons dari API polusi udara
export interface AirPollutionResponse {
  list: {
    main: {
      aqi: number; // Indeks Kualitas Udara (1-5)
    };
    components: {
      co: number;  // Karbon monoksida
      no2: number; // Nitrogen dioksida
      o3: number;  // Ozon
      so2: number; // Sulfur dioksida
    };
  }[];
}
