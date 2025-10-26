import { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';

// Tipe tetap sama
type Theme = 'light' | 'dark';
interface ThemeContextType {
  theme: Theme;
  // --- UBAH INI ---
  // toggleTheme sekarang akan menerima event klik
  toggleTheme: (event: React.MouseEvent) => void; 
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const getInitialTheme = (): Theme => {
  if (typeof window !== 'undefined') {
    const storedTheme = localStorage.getItem('theme') as Theme | null;
    return storedTheme || 'light';
  }
  return 'light';
};

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState<Theme>(getInitialTheme);

  // Effect ini sekarang hanya untuk mengatur tema awal saat load
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, []); // <-- Array kosong berarti ini hanya berjalan sekali saat load

  // --- UBAH FUNGSI INI ---
  const toggleTheme = (event: React.MouseEvent) => {
    const newTheme = theme === 'light' ? 'dark' : 'light';

    // Fallback untuk browser lama (Firefox, Safari)
    // @ts-ignore
    if (!document.startViewTransition) {
      setTheme(newTheme);
      document.documentElement.setAttribute('data-theme', newTheme);
      localStorage.setItem('theme', newTheme);
      return;
    }

    // --- Logika Animasi View Transition ---

    // 1. Dapatkan koordinat klik
    const x = event.clientX;
    const y = event.clientY;

    // 2. Hitung radius terbesar untuk menutupi layar
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y)
    );

    // 3. Simpan koordinat & radius sebagai variabel CSS
    document.documentElement.style.setProperty('--clip-x', `${x}px`);
    document.documentElement.style.setProperty('--clip-y', `${y}px`);
    document.documentElement.style.setProperty('--clip-r', `${endRadius}px`);

    // 4. Mulai transisi
    // @ts-ignore
    document.startViewTransition(() => {
      // Kode ini dieksekusi setelah "snapshot" lama diambil
      
      // A. Perbarui DOM secara sinkron
      document.documentElement.setAttribute('data-theme', newTheme);
      
      // B. Perbarui state React
      setTheme(newTheme);

      // C. Perbarui localStorage
      localStorage.setItem('theme', newTheme);
    });
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};