import { useState } from "react";
import "./App.css";
import { StartServer, StopServer } from "../wailsjs/go/app/App";
import { Notifications } from "./Components/Notifications";
import { LocalIpList } from "./Components/LocalIpList";
import { ClientUpload } from "./Components/ClientUpload";
import { CertificateHash } from "./Components/CertificateHash";
import { Auth } from "./Components/Auth";
import { FilesList } from "./Components/FileList";
import { PinDisplay } from "./Components/PinDisplay";

const SERVER_PORT = 53317;

function App() {
  const [serverRunning, setServerRunning] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const handleServerToggle = async () => {
    try {
      if (serverRunning) {
        await StopServer();
        setServerRunning(false);
      } else {
        await StartServer(SERVER_PORT);
        setServerRunning(true);
      }
    } catch (error) {
      console.error("Failed to start server:", error);
    }
  };

  const handleLoginSuccess = () => {
    setIsAuthenticated(true);
  };

  // If not authenticated, show login screen
  if (!isAuthenticated) {
    return <Auth onLoginSuccess={handleLoginSuccess} />;
  }

  return (
    <div className="container">
      <div className="card">
        <h1>Tella Desktop</h1>
        <LocalIpList />
      </div>
      <div className="card">
        {serverRunning && (
          <div className="server-status running">
            Server is running on port {SERVER_PORT}
          </div>
        )}
        <button
          className={`button ${serverRunning ? "button-success" : "button-primary"}`}
          onClick={handleServerToggle}
        >
          {serverRunning ? "Stop Server" : "Start Server"}
        </button>
        <PinDisplay serverRunning={serverRunning} />
        <CertificateHash serverRunning={serverRunning} />
      </div>

      <ClientUpload />

      {/* New Files List Component */}
      <div className="card">
        <FilesList />
      </div>
      
      <Notifications />
    </div>
  );
}

export default App;
