import { useState, useEffect } from "react";
import { GetLocalIPs } from "../../../wailsjs/go/app/App";

export function LocalIpList() {
    const [localIPs, setLocalIPs] = useState<string[]>([])

    useEffect(() => {
        const fetchIPs = async() => {
            try {
                const ips = await GetLocalIPs()
                setLocalIPs(ips)
            } catch (error) {
                console.error('failed to get local IPs:', error)
            }
        }

        fetchIPs()
    }, [])

    if(localIPs.length > 0) {
        return (
            <div>
                <h3>Local IP Addresses:</h3>
                <ul>
                    {localIPs.map((ip, index) => (
                    <li key={index}>
                        {ip}
                    </li>
                    ))}
                </ul>
            </div>
        )
    } else {
        return null
    }
}