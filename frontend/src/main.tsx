import React from 'react'
import {createRoot} from 'react-dom/client'
import App from './App'
import { ThemeProvider } from './styles'
import { ServerProvider } from './Contexts/ServerContext'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
      <ThemeProvider>
          <ServerProvider>
            <App />
          </ServerProvider>
        </ThemeProvider>
    </React.StrictMode>
)
