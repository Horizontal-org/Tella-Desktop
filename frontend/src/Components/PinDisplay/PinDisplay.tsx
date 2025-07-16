import { useState, useEffect } from 'react';
import { GetServerPIN } from '../../../wailsjs/go/app/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

interface Props {
    serverRunning: boolean;
}

export function PinDisplay({ serverRunning }: Props) {
    const [pin, setPin] = useState('');

    useEffect(() => {
        // Get PIN when component mounts or server starts
        const fetchPIN = async () => {
            if (serverRunning) {
                try {
                    const currentPIN = await GetServerPIN();
                    setPin(currentPIN);
                } catch (error) {
                    console.error('Failed to get PIN:', error);
                    setPin('');
                }
            } else {
                setPin('');
            }
        };
        
        fetchPIN();
        
        // Listen for PIN updates
        const cleanup = EventsOn("server-pin", (newPIN) => {
            setPin(newPIN.toString());
        });
        
        return cleanup;
    }, [serverRunning]);
    
    if (!serverRunning || !pin) {
        return <span>-</span>;
    }
    
    return <span>{pin}</span>;
}