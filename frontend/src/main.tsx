// File: main.tsx

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import { ThemeProvider } from './ThemeContext.tsx' // <-- 1. IMPORT
import './index.css' // (Jika ada)
import './App.css' // <-- Pastikan CSS App juga diimpor

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    {/* 2. PASTIKAN <App /> ADA DI DALAM SINI */}
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </React.StrictMode>,
)