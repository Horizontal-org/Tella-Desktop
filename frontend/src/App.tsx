import {useState} from 'react';
import './App.css';
import { StartServer, StopServer } from '../wailsjs/go/app/App';
import { Notifications } from './Components/Notifications';
import { LocalIpList } from './Components/LocalIpList';
import { ClientUpload } from './Components/ClientUpload'
import { CertificateDisplay } from './Components/CertificateDisplay';

const SERVER_PORT = 53317

function App() {
    const [serverRunning, setServerRunning] = useState(false)

    const handleServerToggle = async () => {
        try {
            if(serverRunning) {
                await StopServer()
                setServerRunning(false)
            } else {
                await StartServer(SERVER_PORT)
                setServerRunning(true)
            }
        } catch (error) {
            console.error('Failed to start server:', error)
        }
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
                    className={`button ${serverRunning ? 'button-success' : 'button-primary'}`}
                    onClick={handleServerToggle}
                >
                    {serverRunning ? 'Stop Server' : 'Start Server'}
                </button>
            </div>

            <CertificateDisplay />
            <ClientUpload />
            <Notifications />
        </div>
    )
}

export default App
