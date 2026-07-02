import styled, { keyframes } from 'styled-components';

interface ErrorDialogProps {
  onClose: () => void;
  text: string;
  buttonLabel: string;
}

export function ErrorDialog({ 
  onClose, text, buttonLabel
}: ErrorDialogProps) {
  return (
    <ModalOverlay>
      <ModalContainer>
        <ModalHeader>
          <Title>Connection failed</Title>
        </ModalHeader>
        
        <Description>
        {text}
        </Description>
        
        
        <ModalFooter>
          <CancelButton onClick={onClose}>
          {buttonLabel}
          </CancelButton>
        </ModalFooter>
      </ModalContainer>
    </ModalOverlay>
  );
}

const ModalOverlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
`;

const ModalContainer = styled.div`
  background-color: white;
  border-radius: 8px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow: hidden;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  padding: 2rem 4rem 2rem 4rem;
`;

const ModalHeader = styled.div`
  padding-top: 1.5rem;
`;

const Title = styled.h2`
  margin: 0;
  color: #212529;
  font-size: 1.5rem;
  font-weight: 600;
`;

const Description = styled.p`
  color: #6c757d;
  margin-bottom: 2rem;
  font-size: 1rem;
  line-height: 1.5;
`;

const ModalFooter = styled.div`
  padding: 1.5rem 0;
  display: flex;
  gap: 1rem;
  justify-content: end;
`;

const Button = styled.button`
  padding: 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  text-transform: uppercase;
  min-width: 80px;
`;

const CancelButton = styled(Button)`
  text-transform: uppcase;
  background-color: white;
  color: #6c757d;
  border: 1px solid #D9D9D9;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;
