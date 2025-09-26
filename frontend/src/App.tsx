import { useState } from "react";
import "./App.css";
import { AppRouter } from "./Router/AppRouter";

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const handleLoginSuccess = () => {
    setIsAuthenticated(true);
  };

  const handleLock = () => {
    setIsAuthenticated(false);
  };

  return (
    <AppRouter
      isAuthenticated={isAuthenticated}
      onLoginSuccess={handleLoginSuccess}
      onLock={handleLock}
    />
  );
}

export default App;
