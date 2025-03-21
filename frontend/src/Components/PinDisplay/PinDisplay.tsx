import { useState, useEffect } from 'react';
import { GetServerPIN } from '../../../wailsjs/go/app/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import styled from 'styled-components';

interface Props {
    serverRunning: boolean;
}

export function PinDisplay({ serverRunning }: Props) {
    const [pin, setPin] = useState('');

    useEffect(() => {
        // Get PIN when component mounts or server starts
        const fetchPIN = async () => {
            if (serverRunning) {
                const currentPIN = await GetServerPIN();
                setPin(currentPIN);
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
        return null;
    }
    
    return (
        <PinContainer>
            <PinLabel>Connection PIN:</PinLabel>
            <PinCode>{pin}</PinCode>
            <PinHint>Share this PIN with the sender to verify the connection</PinHint>
        </PinContainer>
    );
}

const PinContainer = styled.div`
    margin: 1rem 0;
    padding: 1rem;
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: ${({ theme }) => theme.borderRadius.default};
    text-align: center;
`;

const PinLabel = styled.div`
    font-weight: bold;
    margin-bottom: 0.5rem;
    color: ${({ theme }) => theme.colors.lightGray};
`;

const PinCode = styled.div`
    font-size: 2rem;
    font-weight: bold;
    letter-spacing: 0.5rem;
    margin: 0.5rem 0;
    color: ${({ theme }) => theme.colors.darkGray};
`;

const PinHint = styled.div`
    font-size: 0.8rem;
    color: ${({ theme }) => theme.colors.lightGray};
`;