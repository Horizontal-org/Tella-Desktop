import styled from 'styled-components';

interface CertificateVerificationModalProps {
  isOpen: boolean;
  certificateHash: string;
  onConfirm: () => void;
  onDiscard: () => void;
}

export function CertificateVerificationModal({ 
  isOpen, 
  certificateHash, 
  onConfirm, 
  onDiscard 
}: CertificateVerificationModalProps) {
  if (!isOpen) return null;

  const formatHash = (hash: string) => {
    return hash.match(/.{1,4}/g)?.join(' ') || hash;
  };

  return (
    <ModalOverlay>
      <ModalContainer>
        <ModalHeader>
          <Title>Verification</Title>
        </ModalHeader>
        
        <ModalContent>
          <Description>
            Make sure that this sequence matches what is shown on the sender's device.
          </Description>
          
          <HashContainer>
            <HashText>{formatHash(certificateHash)}</HashText>
          </HashContainer>
          
          <Warning>
            If the sequence on your device does not match the sequence on the sender's device, 
            the connection may not be secure and should be discarded.
          </Warning>
        </ModalContent>
        
        <ModalFooter>
          <DiscardButton onClick={onDiscard}>
            DISCARD AND START OVER
          </DiscardButton>
          <ConfirmButton onClick={onConfirm}>
            CONFIRM AND CONNECT
          </ConfirmButton>
        </ModalFooter>
      </ModalContainer>
    </ModalOverlay>
  );
}

// Styled Components
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
`;

const ModalHeader = styled.div`
  padding: 1.5rem;
  border-bottom: 1px solid #e9ecef;
  text-align: center;
`;

const Title = styled.h2`
  margin: 0;
  color: #212529;
  font-size: 1.5rem;
  font-weight: 600;
`;

const ModalContent = styled.div`
  padding: 2rem 1.5rem;
  text-align: center;
`;

const Description = styled.p`
  color: #6c757d;
  margin-bottom: 2rem;
  font-size: 1rem;
  line-height: 1.5;
`;

const HashContainer = styled.div`
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
`;

const HashText = styled.code`
  font-family: 'Courier New', monospace;
  font-size: 1rem;
  color: #212529;
  word-break: break-all;
  line-height: 1.6;
`;

const Warning = styled.p`
  color: #6c757d;
  font-size: 0.9rem;
  line-height: 1.5;
  margin: 0;
`;

const ModalFooter = styled.div`
  padding: 1.5rem;
  border-top: 1px solid #e9ecef;
  display: flex;
  gap: 1rem;
  justify-content: center;
`;

const Button = styled.button`
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  text-transform: uppercase;
  min-width: 140px;
`;

const DiscardButton = styled(Button)`
  background-color: white;
  color: #6c757d;
  border: 1px solid #6c757d;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;

const ConfirmButton = styled(Button)`
  background-color: #28a745;
  color: white;
  border: 1px solid #28a745;
  
  &:hover {
    background-color: #218838;
  }
`;