import React, { useState, useEffect } from 'react';
import { VerifyPassword } from '../../../wailsjs/go/app/App';
import { 
  AuthContainer, 
  AuthCard, 
  CardTitle,
  CardSubtitle, 
  FormGroup, 
  Label, 
  Input, 
  AuthButton, 
  ErrorMessage 
} from './styles';

interface LoginProps {
  onLoginSuccess: () => void;
  initialError?: string;
}

export function Login({ onLoginSuccess, initialError = '' }: LoginProps) {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (initialError) {
      setError(initialError);
    }
  }, [initialError]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (!password) {
      setError('Please enter your password');
      return;
    }

    setLoading(true);
    try {
      await VerifyPassword(password);
      onLoginSuccess();
    } catch (error: any) {
      setError('Invalid password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthContainer>
      <AuthCard>
        <CardTitle>Tella</CardTitle>
        <CardSubtitle>Enter your password to log in</CardSubtitle>
        
        {error && <ErrorMessage>{error}</ErrorMessage>}
        
        <form onSubmit={handleLogin}>
          <FormGroup>
            <Label htmlFor="password">Password</Label>
            <Input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter password"
              disabled={loading}
            />
          </FormGroup>
          
          <AuthButton 
            type="submit" 
            disabled={loading}
          >
            {loading ? 'Loading...' : 'LOG IN'}
          </AuthButton>
        </form>
      </AuthCard>
    </AuthContainer>
  );
}