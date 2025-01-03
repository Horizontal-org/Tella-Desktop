package network

import "net"

func GetLocalIPs() ([]string, error) {
    var ips []string
    
    interfaces, err := net.Interfaces()
    if err != nil {
        return nil, err
    }

    for _, iface := range interfaces {
        if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
            continue
        }

        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }

        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            
            if ip == nil || ip.IsLoopback() {
                continue
            }
            
            ip = ip.To4()
            if ip == nil {
                continue
            }
            
            ips = append(ips, ip.String())
        }
    }
    
    return ips, nil
}