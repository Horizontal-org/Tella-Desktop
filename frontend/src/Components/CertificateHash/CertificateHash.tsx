// frontend/src/Components/CertificateHash/CertificateHash.tsx
import React, { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

interface Props {
    serverRunning: boolean;
}

export function CertificateHash({ serverRunning }: Props) {
    const [certHash, setCertHash] = useState('');

    useEffect(() => {
        console.log("Subscribing to certificate-hash events");

        const cleanup = EventsOn("certificate-hash", (data) => {
            console.log("Received certificate hash event:", data);
            setCertHash(data.toString());
        });

        return () => {
            console.log("Cleaning up certificate-hash subscription");
            cleanup();
        };
    }, []);

    if (!serverRunning || !certHash) {
        return null;
    }

    return (
        <div className="status">
            <h3>Certificate Hash:</h3>
            <code>{certHash.match(/.{1,4}/g)?.join(' ')}</code>
        </div>
    );
}