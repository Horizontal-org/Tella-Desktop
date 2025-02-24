// frontend/src/Components/Auth/Login.tsx
import React, { useState, useEffect } from 'react';
import { IsFirstTimeSetup, CreatePassword, VerifyPassword } from '../../../wailsjs/go/app/App';
import './Auth.css';

interface LoginProps {
  onLoginSuccess: () => void;
}

export function Login({ onLoginSuccess }: LoginProps) {
  const [isFirstTime, setIsFirstTime] = useState<boolean | null>(null);
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

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

  const handleCreatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    // Basic validation
    if (password.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setLoading(true);
    try {
      await CreatePassword(password);
      onLoginSuccess();
    } catch (error: any) {
      setError(error.toString());
    } finally {
      setLoading(false);
    }
  };

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

  // Show loading while checking if it's first time setup
  if (isFirstTime === null) {
    return <div className="auth-container"><div className="loading">Loading...</div></div>;
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>{isFirstTime ? 'Create Password' : 'Login'}</h1>
        
        {error && <div className="error-message">{error}</div>}
        
        <form onSubmit={isFirstTime ? handleCreatePassword : handleLogin}>
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter password"
              disabled={loading}
            />
          </div>
          
          {isFirstTime && (
            <div className="form-group">
              <label htmlFor="confirmPassword">Confirm Password</label>
              <input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm password"
                disabled={loading}
              />
            </div>
          )}
          
          <button 
            type="submit" 
            className="auth-button" 
            disabled={loading}
          >
            {loading ? 'Loading...' : isFirstTime ? 'Create Password' : 'Login'}
          </button>
        </form>
      </div>
    </div>
  );
}