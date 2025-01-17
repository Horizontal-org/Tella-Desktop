import React, { useState } from 'react';
import { RegisterWithDevice, SendTestFile, VerifyServerCertificate } from '../../../wailsjs/go/app/App';

export function ClientUpload() {
    const [ip, setIp] = useState('');
    const [port, setPort] = useState('53317');
    const [pin, setPin] = useState('');
    const [status, setStatus] = useState('');
    const [loading, setLoading] = useState(false);
    const [certificateVerified, setCertificateVerified] = useState(false);


    const handleVerifyCertificate = async () => {
        try {
            setLoading(true);
            setStatus('Verifying server certificate...');
            await VerifyServerCertificate(ip, parseInt(port));
            setCertificateVerified(true);
            setStatus('Certificate received. Please verify the fingerprint before proceeding.');
        } catch (error) {
            setStatus(`Certificate verification failed: ${error}`);
            setCertificateVerified(false);
        } finally {
            setLoading(false);
        }
    };

    const handleRegister = async () => {
        try {
            setLoading(true);
            setStatus('Registering...');
            await RegisterWithDevice(ip, parseInt(port));
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
                        onClick={handleVerifyCertificate}
                        disabled={loading || !ip}
                        className="button button-primary"
                    >
                        Verify Certificate
                    </button>
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
            </div>
        </div>
    )
}