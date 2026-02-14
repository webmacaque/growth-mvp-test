import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import './index.css'
import App from './App.tsx'
import TelegramIntegrationPage from './pages/TelegramIntegrationPage.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<App />}>
          <Route path="/shops/:shopId/growth/telegram" element={<TelegramIntegrationPage />} />
          <Route path="*" element={<Navigate to="/shops/1/growth/telegram" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
