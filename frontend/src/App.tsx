import { useState } from "react";
import "./App.css";
import { AppRouter } from "./Router/AppRouter";

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const handleLoginSuccess = () => {
    setIsAuthenticated(true);
  };


  return (
    <AppRouter isAuthenticated={isAuthenticated} onLoginSuccess={handleLoginSuccess}/>
  );
}

export default App;
