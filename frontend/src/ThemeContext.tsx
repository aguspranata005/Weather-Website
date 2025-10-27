// Impor 'type' ditambahkan untuk ReactNode dan MouseEvent
import { createContext, useContext, useState, useEffect, type ReactNode, type MouseEvent } from 'react';

// Tipe tetap sama
type Theme = 'light' | 'dark';
interface ThemeContextType {
  theme: Theme;
  toggleTheme: (event: MouseEvent) => void; 
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const getInitialTheme = (): Theme => {
  // Langsung kembalikan 'light' sebagai default, tanpa memeriksa localStorage atau preferensi sistem.
  return 'light';
};

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState<Theme>(getInitialTheme);

  // Effect ini sekarang menjadi SATU-SATUNYA sumber kebenaran.
  // Ini berjalan saat 'theme' berubah, memperbarui DOM dan localStorage.
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]); // <-- DEPENDENSI [theme] SANGAT PENTING

  const toggleTheme = (event: MouseEvent) => {
    const newTheme = theme === 'light' ? 'dark' : 'light';

    // @ts-ignore - document.startViewTransition
    if (!document.startViewTransition) {
      setTheme(newTheme); // Cukup update state, useEffect akan menangani DOM/localStorage
      return;
    }

    // --- Logika Animasi View Transition ---
    const x = event.clientX;
    const y = event.clientY;
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y)
    );

    document.documentElement.style.setProperty('--clip-x', `${x}px`);
    document.documentElement.style.setProperty('--clip-y', `${y}px`);
    document.documentElement.style.setProperty('--clip-r', `${endRadius}px`);

    // @ts-ignore - document.startViewTransition
    document.startViewTransition(() => {
      // CUKUP PERBARUI STATE REACT.
      // useEffect yang sudah kita perbaiki akan menangani sisanya.
      setTheme(newTheme);
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