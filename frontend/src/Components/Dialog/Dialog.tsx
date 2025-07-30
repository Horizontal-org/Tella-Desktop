import React from 'react';
import styled from 'styled-components';

interface DialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  confirmButtonText?: string;
  cancelButtonText?: string;
  children: React.ReactNode;
}

const Overlay = styled.div<{ $isOpen: boolean }>`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: ${props => props.$isOpen ? 'flex' : 'none'};
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

const DialogContainer = styled.div`
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow: hidden;
  animation: slideIn 0.2s ease-out;

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
`;

const DialogHeader = styled.div`
  padding: 1.5rem 1.5rem 1rem 1.5rem;
  border-bottom: 1px solid #e9ecef;
`;

const DialogTitle = styled.h2`
  font-size: 1.25rem;
  font-weight: 600;
  color: #333;
  margin: 0;
`;

const DialogBody = styled.div`
  padding: 1.5rem;
  color: #6c757d;
  line-height: 1.6;
`;

const DialogFooter = styled.div`
  padding: 1.5rem;
  border-top: 1px solid #e9ecef;
  display: flex;
  gap: 1rem;
  justify-content: end;
`;

const Button = styled.button`
  background-color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  text-transform: uppercase;
  min-width: 140px;
  border: 1px solid #6c757d;
  color: #6c757d;
`;

export function Dialog({ 
  isOpen, 
  onClose, 
  onConfirm, 
  title, 
  confirmButtonText = 'CONFIRM',
  cancelButtonText = 'CANCEL',
  children 
 }: DialogProps) {
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
    if (e.key === 'Enter') {
      onConfirm();
    }
  };

  return (
    <Overlay 
      $isOpen={isOpen} 
      onClick={handleOverlayClick}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
    >
      <DialogContainer>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        
        <DialogBody>
          {children}
        </DialogBody>
        
        <DialogFooter>
          <Button onClick={onClose}>
            {cancelButtonText}
          </Button>
          <Button
            onClick={onConfirm}
          >
            {confirmButtonText}
          </Button>
        </DialogFooter>
      </DialogContainer>
    </Overlay>
  );
}