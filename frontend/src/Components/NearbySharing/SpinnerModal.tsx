import styled, { keyframes } from 'styled-components';

interface SpinnerModalProps {
  isOpen: boolean;
  onCancel: () => void;
}

export function SpinnerModal({ 
  isOpen, 
  onCancel 
}: SpinnerModalProps) {
  if (!isOpen) return null;

  return (
    <ModalOverlay>
      <ModalContainer>
        <ModalHeader>
          <Title>Verification</Title>
        </ModalHeader>
        
        <Description>
        Waiting for the sender to also confirm.
        </Description>
        
        <SpinnerContainer>
	    <svg width="24" height="24" stroke="#A9A9A9" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
	        <SpinnerG>
	            <SpinnerCircle cx="12" cy="12" r="9.5" fill="none" strokeWidth="2"/>
	        </SpinnerG>
	    </svg>
        </SpinnerContainer>
        
        <ModalFooter>
          <CancelButton onClick={onCancel}>
          CANCEL
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
  background-color: white;
  color: #6c757d;
  border: 1px solid #D9D9D9;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;

const SpinnerContainer = styled.div`
    text-align: center;
`;

const spinnerAnimOne = keyframes`
    100% {transform:rotate(360deg)} 
`;

const spinnerAnimTwo = keyframes`
    0% {stroke-dasharray:0 150;stroke-dashoffset:0} 
    47.5% {stroke-dasharray:42 150;stroke-dashoffset:-16} 
    95%,100%{stroke-dasharray:42 150;stroke-dashoffset:-59}
`;

const SpinnerG = styled.g`
    transform-origin:center;
    animation:${spinnerAnimOne} 2s linear infinite
`;

const SpinnerCircle =  styled.circle`
    stroke-linecap:round;animation:${spinnerAnimTwo} 1.5s ease-in-out infinite
`;

