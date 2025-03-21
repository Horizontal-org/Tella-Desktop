import { useState, useEffect } from 'react';
import { IsFirstTimeSetup } from '../../../wailsjs/go/app/App';
import { SignUp } from './SignUp';
import { Login } from './Login'
import { AuthContainer, LoadingMessage } from './styles';

interface AuthProps {
  onLoginSuccess: () => void;
}

export function Auth({ onLoginSuccess }: AuthProps) {
  const [isFirstTime, setIsFirstTime] = useState<boolean | null>(null);
  const [error, setError] = useState('');

  // Check if it's first time setup on component mount
  useEffect(() => {
    const checkFirstTimeSetup = async () => {
      try {
        const isFirst = await IsFirstTimeSetup();
        setIsFirstTime(isFirst);
      } catch (error) {
        console.error('Failed to check if first time setup:', error);
        setError('Failed to initialize application. Please restart.');
      }
    };

    checkFirstTimeSetup();
  }, []);

  // Show loading while checking if it's first time setup
  if (isFirstTime === null) {
    return (
      <AuthContainer>
        <LoadingMessage>Loading...</LoadingMessage>
      </AuthContainer>
    );
  }

  if (isFirstTime) {
    return <SignUp onLoginSuccess={onLoginSuccess} initialError={error} />;
  } else {
    return <Login onLoginSuccess={onLoginSuccess} initialError={error} />;
  }
}