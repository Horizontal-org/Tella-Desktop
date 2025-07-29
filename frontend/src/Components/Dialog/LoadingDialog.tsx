import React from 'react';
import styled, { keyframes } from 'styled-components';

interface LoadingDialogProps {
  isOpen: boolean;
  onCancel: () => void;
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
  text-align: center;
`;

const DialogBody = styled.div`
  padding: 3rem 2rem;
`;

const DialogTitle = styled.h2`
  font-size: 1.5rem;
  font-weight: 600;
  color: #333;
  margin: 0 0 1.5rem 0;
`;

const DialogMessage = styled.p`
  color: #6c757d;
  line-height: 1.6;
  margin: 0 0 2rem 0;
`;

const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

const Spinner = styled.div`
  width: 32px;
  height: 32px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #007bff;
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
  margin: 0 auto 2rem auto;
`;

const DialogFooter = styled.div`
  padding: 1.5rem;
  border-top: 1px solid #e9ecef;
  display: flex;
  justify-content: center;
`;

const CancelButton = styled.button`
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  text-transform: uppercase;
  min-width: 140px;
  background-color: white;
  color: #6c757d;
  border: 1px solid #6c757d;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;

export function LoadingDialog({ isOpen, onCancel }: LoadingDialogProps) {
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onCancel();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onCancel();
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
        <DialogBody>
          <DialogTitle>Your files are exporting</DialogTitle>
          <DialogMessage>
            Please wait while your files are exporting. Do not close Tella Desktop or the export may fail.
          </DialogMessage>
          <Spinner />
        </DialogBody>
        
        <DialogFooter>
          <CancelButton onClick={onCancel}>
            Cancel
          </CancelButton>
        </DialogFooter>
      </DialogContainer>
    </Overlay>
  );
}