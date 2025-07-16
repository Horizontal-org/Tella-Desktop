import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Auth } from '../Components/Auth';
import { Dashboard } from '../Components/Dashboard';
import { ProtectedRoute } from '../Components/ProtectedRoute/ProtectedRoute';
import { NearbySharing } from '../Components/NearbySharing/NearbySharing';

interface AppRouterProps {
  isAuthenticated: boolean;
  onLoginSuccess: () => void;
}

export function AppRouter({ isAuthenticated, onLoginSuccess }: AppRouterProps) {
  return (
    <BrowserRouter>
      <Routes>
        <Route 
          path="/auth" 
          element={
            isAuthenticated ? (
              <Navigate to="/" replace />
            ) : (
              <Auth onLoginSuccess={onLoginSuccess} />
            )
          } 
        />

        <Route 
          path="/" 
          element={
            <ProtectedRoute isAuthenticated={isAuthenticated}>
              <Dashboard />
            </ProtectedRoute>
          } 
        />

        <Route 
          path="/nearby-sharing" 
          element={
            <ProtectedRoute isAuthenticated={isAuthenticated}>
              <NearbySharing />
            </ProtectedRoute>
          } 
        />
        
        
        <Route 
          path="*" 
          element={
            <Navigate 
              to={isAuthenticated ? "/" : "/auth"} 
              replace 
            />
          } 
        />
      </Routes>
    </BrowserRouter>
  );
}