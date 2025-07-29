import React, { useEffect } from 'react';
import styled, { keyframes } from 'styled-components';

interface SuccessToastProps {
  isVisible: boolean;
  message: string;
  onClose: () => void;
  autoCloseMs?: number;
}

const slideIn = keyframes`
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
`;

const slideOut = keyframes`
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
`;

const ToastContainer = styled.div<{ $isVisible: boolean }>`
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background-color: #28a745;
  color: white;
  padding: 1rem 2rem;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 2000;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 300px;
  max-width: 500px;
  animation: ${props => props.$isVisible ? slideIn : slideOut} 0.3s ease-out;
  display: ${props => props.$isVisible ? 'flex' : 'none'};
`;

const CheckIcon = styled.div`
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background-color: white;
  color: #28a745;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.875rem;
  flex-shrink: 0;
`;

const Message = styled.span`
  font-weight: 500;
  flex: 1;
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0;
  margin: 0;
  font-size: 1.2rem;
  line-height: 1;
  opacity: 0.8;
  transition: opacity 0.2s;
  
  &:hover {
    opacity: 1;
  }
`;

export function SuccessToast({ 
  isVisible, 
  message, 
  onClose, 
  autoCloseMs = 5000 
}: SuccessToastProps) {
  useEffect(() => {
    if (isVisible && autoCloseMs > 0) {
      const timer = setTimeout(onClose, autoCloseMs);
      return () => clearTimeout(timer);
    }
  }, [isVisible, onClose, autoCloseMs]);

  if (!isVisible) return null;

  return (
    <ToastContainer $isVisible={isVisible}>
      <CheckIcon>✓</CheckIcon>
      <Message>{message}</Message>
      <CloseButton onClick={onClose}>×</CloseButton>
    </ToastContainer>
  );
}