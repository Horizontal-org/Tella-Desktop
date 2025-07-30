import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Auth } from '../Components/Auth';
import { ProtectedRoute } from '../Components/ProtectedRoute/ProtectedRoute';
import { NearbySharing } from '../Components/NearbySharing/NearbySharing';
import { FileList } from '../Components/FileList/FileList';
import { AppLayout } from '../Components/AppLayout/AppLayout';
import { FolderList } from '../Components/FolderList';
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
          path="/*" 
          element={
            <ProtectedRoute isAuthenticated={isAuthenticated}>
              <AppLayout>
                <Routes>
                  <Route path="/" element={<FolderList />} />
                  <Route path="/folder/:folderId" element={<FileList />} />
                  <Route path="/nearby-sharing" element={<NearbySharing />} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </AppLayout>
            </ProtectedRoute>
          } 
        />
      </Routes>
    </BrowserRouter>
  );
}