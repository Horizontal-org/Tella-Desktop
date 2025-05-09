import React, { useState, useEffect } from 'react';
import { RegisterWithDevice, SendTestFile } from '../../../wailsjs/go/app/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

export function ClientUpload() {
    const [ip, setIp] = useState('');
    const [port, setPort] = useState('53317');
    const [pin, setPin] = useState('');
    const [status, setStatus] = useState('');
    const [loading, setLoading] = useState(false);
    const [certHash, setCertHash] = useState('');

    useEffect(() => {
        const cleanup = EventsOn("client-certificate-hash", (data) => {
            console.log("Client received certificate hash:", data);
            setCertHash(data.toString());
        });

        return () => cleanup();
    }, []);

    const handleRegister = async () => {
        try {
            setLoading(true);
            setCertHash('');
            setStatus('Registering...');
            await RegisterWithDevice(ip, parseInt(port), pin);
            setStatus('Successfully registered with device');
        } catch (error) {
            setStatus(`Registration failed: ${error}`);
        } finally {
            setLoading(false);
        }
    };

    const handleSendFile = async () => {
        try {
            setLoading(true);
            setStatus('Sending file...');
            await SendTestFile(ip, parseInt(port), pin);
            setStatus('File sent successfully');
        } catch (error) {
            setStatus(`Failed to send file: ${error}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="card">
            <h2>Connect to Device</h2>
            
            <div>
                <div>
                    <label className="label">IP Address:</label>
                    <input
                        type="text"
                        value={ip}
                        onChange={(e) => setIp(e.target.value)}
                        className="input"
                        placeholder="192.168.1.x"
                    />
                </div>

                <div>
                    <label className="label">Port:</label>
                    <input
                        type="text"
                        value={port}
                        onChange={(e) => setPort(e.target.value)}
                        className="input"
                    />
                </div>

                <div>
                    <label className="label">PIN (optional):</label>
                    <input
                        type="text"
                        value={pin}
                        onChange={(e) => setPin(e.target.value)}
                        className="input"
                        placeholder="Enter PIN"
                    />
                </div>

                <div className="button-group">
                    <button 
                        onClick={handleRegister}
                        disabled={loading || !ip}
                        className="button button-primary"
                    >
                        Register
                    </button>
                    <button 
                        onClick={handleSendFile}
                        disabled={loading || !ip}
                        className="button button-success"
                    >
                        Send Test File
                    </button>
                </div>

                {status && (
                    <div className={`status ${status.includes('failed') ? 'status-error' : 'status-success'}`}>
                        {status}
                    </div>
                )}

                {certHash && (
                    <div className="status status-success">
                        <h3>Remote Certificate Hash:</h3>
                        <code>{certHash.match(/.{1,4}/g)?.join(' ')}</code>
                    </div>
                )}

            </div>
        </div>
    )
}