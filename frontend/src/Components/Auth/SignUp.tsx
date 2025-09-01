import React, { useState, useEffect } from "react";
import { CreatePassword } from "../../../wailsjs/go/app/App";
import { 
  AuthContainer, 
  AuthCard, 
  CardTitle, 
  FormGroup, 
  Label, 
  Input, 
  AuthButton, 
  ErrorMessage, 
  CardSubtitle
} from './styles';

interface SignUpProps {
  onLoginSuccess: () => void;
  initialError?: string;
}

export function SignUp({ onLoginSuccess, initialError = '' }: SignUpProps) {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (initialError) {
      setError(initialError);
    }
  }, [initialError]);

  const handleCreatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Basic validation
    if (password.length < 6) {
      setError("Password must be at least 6 characters long");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match");
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

  return (
    <AuthContainer>
      <AuthCard>
        <CardTitle>Welcome to Tella</CardTitle>
        <CardSubtitle>
          Create a password to log into Tella and access your files. Your password must be at least 8 characters long.
        </CardSubtitle>
        <CardSubtitle>
          Make sure to store your password in a safe place. If you lose your password, there is no way of recovering your files.
        </CardSubtitle>

        {error && <ErrorMessage>{error}</ErrorMessage>}

        <form onSubmit={handleCreatePassword}>
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

          <FormGroup>
            <Label htmlFor="confirmPassword">Confirm Password</Label>
            <Input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm password"
              disabled={loading}
            />
          </FormGroup>

          <AuthButton type="submit" disabled={loading}>
            {loading ? "Loading..." : "SAVE"}
          </AuthButton>
        </form>
      </AuthCard>
    </AuthContainer>
  );
}
