import {useState} from 'react';
import './App.css';
import { StartServer, StopServer } from '../wailsjs/go/app/App';
import { Notifications } from './Components/Notifications';

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
        <div>
            <div>
                {serverRunning && <p>Server is running on port 53317</p>}
            </div>
            <button onClick={handleServerToggle}>
                {serverRunning ? 'Stop Server' : 'Start Server'}
            </button>

            <Notifications />
        </div>
    )
}

export default App
