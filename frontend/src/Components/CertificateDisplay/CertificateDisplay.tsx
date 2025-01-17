import React, { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

export function CertificateDisplay() {
  const [serverCertInfo, setServerCertInfo] = useState('');
  const [clientCertInfo, setClientCertInfo] = useState('');

  useEffect(() => {
    // Listen for certificate info from server
    const serverCertUnsubscribe = EventsOn('certificate-info', (info) => {
      setServerCertInfo(info);
    });

    // Listen for certificate info from client
    const clientCertUnsubscribe = EventsOn('server-certificate', (info) => {
      setClientCertInfo(info);
    });

    return () => {
      serverCertUnsubscribe();
      clientCertUnsubscribe();
    };
  }, []);

  return (
    <>
      {serverCertInfo && (
        <div className="card">
          <h3>Server Certificate</h3>
          <div className="cert-info">
            <pre>{serverCertInfo}</pre>
          </div>
        </div>
      )}
      
      {clientCertInfo && (
        <div className="card">
          <h3>Client Received Certificate</h3>
          <div className="cert-info">
            <pre>{clientCertInfo}</pre>
          </div>
        </div>
      )}
      
      {(serverCertInfo || clientCertInfo) && (
        <p className="text-warning">
          Important: Verify these fingerprints match before proceeding
        </p>
      )}
    </>
  );
}