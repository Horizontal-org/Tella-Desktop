export const theme = {
    colors: {
      primary: '#2563eb',
      secondary: '#059669',
      background: '#ffffff',
      lightGray: '#5F6368',
      darkGray: '#404040',
      error: '#EF4444',
      success: '#34d399',
    },
    fonts: {
      body: '"Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif',
    },
    fontSizes: {
      small: '0.875rem',
      medium: '1rem',
      large: '1.25rem',
      xlarge: '1.5rem',
      xxlarge: '2rem',
    },
    spacing: {
      xs: '0.25rem',
      sm: '0.5rem',
      md: '1rem',
      lg: '1.5rem',
      xl: '2rem',
    },
    borderRadius: {
      small: '0.25rem',
      default: '0.375rem',
      large: '0.5rem',
    },
    shadows: {
      small: '0 1px 2px rgba(0, 0, 0, 0.05)',
      medium: '0 4px 6px rgba(0, 0, 0, 0.1)',
      large: '0 10px 15px rgba(0, 0, 0, 0.1)',
    },
  };
  
  export type Theme = typeof theme;