// frontend/src/Components/CertificateHash/CertificateHash.tsx
import React, { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

interface Props {
    serverRunning: boolean;
}

function formatHash(hashString: string): string {
    const input = [];
    for (let i = 0; i <= hashString.length; i += 4) {
        // grab groups of 4 characters each
        let entry = hashString.slice(i, i+4) + " ";
        // output a newline after four groups of 4
        if (((i+4) % 16) === 0) {
            entry += "\n";
        }
        input.push(entry);
    }
    // trim the last newline and return as a string
    return input.join("").trim();
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
            <pre><code>{formatHash(certHash)}</code></pre>
        </div>
    );
}
