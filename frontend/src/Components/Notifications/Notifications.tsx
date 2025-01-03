import { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

export function Notifications() {
  const [notification, setNotification] = useState('');
  const [showNotification, setShowNotification] = useState(false);

  useEffect(() => {
    // Listen for device registration events
    const unsubscribe = EventsOn('device-registered', (deviceInfo) => {
      setNotification(`Device registered: ${deviceInfo}`);
      setShowNotification(true);

      setTimeout(() => {
        setShowNotification(false);
      }, 3000);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  if (!showNotification) return null;

  return (
    <div>
      <div>
        <h3>New Device</ h3>
        <p>
          {notification}
        </p>
      </div>
    </div>
  );
}