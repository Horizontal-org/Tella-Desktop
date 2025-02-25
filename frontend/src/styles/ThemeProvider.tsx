// frontend/src/styles/ThemeProvider.tsx
import React from 'react';
import { ThemeProvider as StyledThemeProvider, createGlobalStyle } from 'styled-components';
import { theme } from './theme';

const GlobalStyle = createGlobalStyle`
  html {
    background-color: ${({ theme }) => theme.colors.background};
    color: ${({ theme }) => theme.colors.lightGray};
  }

  body {
    margin: 0;
    font-family: ${({ theme }) => theme.fonts.body};
  }

  #app {
    height: 100vh;
    text-align: center;
  }

  a {
    color: ${({ theme }) => theme.colors.primary};
    text-decoration: none;
  }

  * {
    box-sizing: border-box;
  }
`;

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  return (
    <StyledThemeProvider theme={theme}>
      <GlobalStyle />
      {children}
    </StyledThemeProvider>
  );
};