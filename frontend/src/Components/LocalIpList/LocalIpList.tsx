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
                <h2>Local IP Addresses:</h2>
                <ul className="list">
                    {localIPs.map((ip, index) => (
                        <li key={index} className="list-item">
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