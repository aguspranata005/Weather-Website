// functions/env.d.ts

// 1. Definisikan Environment Variables Anda
interface Env {
  OPENWEATHER_API_KEY: string;
}

// 2. Definisikan tipe 'context' yang diterima
interface EventContext<E, P, Data> {
  request: Request;
  env: E;
  // Anda bisa menambahkan properti lain jika perlu, e.g., params, waitUntil
}

// 3. Definisikan tipe 'PagesFunction' yang hilang
type PagesFunction<
  E = unknown,
  P = unknown,
  Data = unknown
> = (
  context: EventContext<E, P, Data>
) => Response | Promise<Response>;