// functions/api/weather.ts

export const onRequestGet: PagesFunction<Env> = async (context) => {
  // Ambil API key dari environment variable
  const API_KEY = context.env.OPENWEATHER_API_KEY;

  // Ambil parameter 'lat' dan 'lon' dari URL
  const url = new URL(context.request.url);
  const lat = url.searchParams.get('lat');
  const lon = url.searchParams.get('lon');

  if (!lat || !lon) {
    return new Response(JSON.stringify({ error: "Parameter 'lat' dan 'lon' dibutuhkan" }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Buat URL API OpenWeatherMap (sama seperti di backend Go Anda)
  const apiURL = `https://api.openweathermap.org/data/2.5/forecast?lat=${lat}&lon=${lon}&appid=${API_KEY}&units=metric&lang=id`;

  try {
    const response = await fetch(apiURL);
    if (!response.ok) {
      throw new Error(`Error dari OpenWeatherMap: ${response.statusText}`);
    }

    const data = await response.json();

    // Kembalikan data sebagai respons
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (err: any) {
    return new Response(JSON.stringify({ error: err.message }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}