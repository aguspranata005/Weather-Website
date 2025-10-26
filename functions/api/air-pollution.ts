// functions/api/air-pollution.ts

export const onRequestGet: PagesFunction<Env> = async (context) => {
  const API_KEY = context.env.OPENWEATHER_API_KEY;

  const url = new URL(context.request.url);
  const lat = url.searchParams.get('lat');
  const lon = url.searchParams.get('lon');

  if (!lat || !lon) {
    return new Response(JSON.stringify({ error: "Parameter 'lat' dan 'lon' dibutuhkan" }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // URL untuk Air Pollution API
  const apiURL = `https://api.openweathermap.org/data/2.5/air_pollution?lat=${lat}&lon=${lon}&appid=${API_KEY}`;

  try {
    const response = await fetch(apiURL);
    if (!response.ok) {
      throw new Error(`Error dari OpenWeatherMap: ${response.statusText}`);
    }

    const data = await response.json();

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