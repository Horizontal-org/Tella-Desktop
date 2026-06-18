import styled from 'styled-components';
import { formatHash } from "../../util/util"
import { SpinnerModal } from "../NearbySharing/SpinnerModal";
import { ErrorDialog } from "../NearbySharing/ErrorDialog";

// TODO (2026-06-17): implement waitingState below
// confirm receiver | confirm sender
// waitingState = (confirm_receiver && !SenderConfirmedReceiver)

interface NearbySharingError {
    text: string;
    button: string;
    hasError: boolean;
}

interface CertificateVerificationModalProps {
  isOpen: boolean;
  receiverCertificateHash: string;
  senderCertificateHash: string;
  senderConfirmedReceiver: boolean;
  nearbySharingError: NearbySharingError;
  modalState: 'CONFIRM_RECEIVER' | 'WAITING_FOR_SENDER_CONFIRM_RECEIVER' | 'CONFIRM_SENDER' | 'WAITING_FOR_SENDER_CONFIRM_SENDER'; 
  onConfirmReceiverHash: () => void;
  onConfirmSenderHash: () => void;
  onDiscard: () => void;
}

// TODO (2026-06-16):
// confirmReceiverHash -> waiting -> waiting for register request to come in
// confirmSenderHash -> waiting -> send register response (?) <-- current onConfirm
//
export function CertificateVerificationModal({ 
  isOpen, 
  receiverCertificateHash, 
  senderCertificateHash, 
  senderConfirmedReceiver,
  nearbySharingError,
  modalState,
  onDiscard,
  onConfirmReceiverHash,
  onConfirmSenderHash
}: CertificateVerificationModalProps) {
  if (!isOpen) return null;

  const getHashForVerification = () => {
      if (modalState === "CONFIRM_RECEIVER") { return receiverCertificateHash }
      if (modalState === "CONFIRM_SENDER") { return senderCertificateHash }
      return ""
  }

  const isWaitingForSender = (
      modalState === "CONFIRM_SENDER" && !senderConfirmedReceiver
  )

  const getStepTitle = () => { 
      if (modalState === "CONFIRM_RECEIVER") { return "Step 1: Confirm recipient hash" }
      if (modalState === "CONFIRM_SENDER") { return "Step 2: Confirm sender hash" }
      return "Step X: Confirm Y"
  }

  // TODO (2026-06-17): no timeout or error graphic is triggered currently
  // TODO (2026-06-18): do not display step 1 after waiting has been started (or step 2 is already being displayed)
  if (nearbySharingError.hasError) {
      return (
          <ErrorDialog 
           text={nearbySharingError.text} 
           buttonLabel={nearbySharingError.button} 
           onClose={onDiscard}
          />
      )
  }
  if (isWaitingForSender) {
      return (
      <SpinnerModal
        onCancel={onDiscard}
      />
      )
  }
  return (
    <ModalOverlay>
      <ModalContainer>
        <ModalHeader>
          <Title>Verification</Title>
        </ModalHeader>
        
        <Description> 
        {getStepTitle()}
        </Description>

        <HashContainer $isSender={modalState === "CONFIRM_SENDER"}>
        <pre>
        <HashText $isSender={modalState === "CONFIRM_SENDER"}>{formatHash(getHashForVerification())}</HashText>
        </pre>
        </HashContainer>

        <Warning>
            Make sure that this sequence matches what is shown on the sender's device.
        </Warning>
        <Warning>
            If the sequence on your device does not match the sequence on the sender's device, the connection may not be secure and should be discarded.
        </Warning>

        <ModalFooter>
          <DiscardButton onClick={onDiscard}>
            DISCARD AND START OVER
          </DiscardButton>
          
          { modalState === "CONFIRM_RECEIVER" ? (
              <ConfirmButton onClick={onConfirmReceiverHash}>
              CONFIRM AND CONTINUE
              </ConfirmButton>
          )
              : (
                  <ConfirmButton onClick={onConfirmSenderHash}>
                  CONFIRM AND CONNECT
                  </ConfirmButton>
              )
          }
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
  max-width: 550px;
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
  color: #5F6368;
  font-weight: 600;
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


// TODO (2026-06-17): dark bg for sender hash verif
const HashContainer = styled.div<{ $isSender?: boolean }>`
  background-color: ${props => props.$isSender ? "#071013CC" : "#D9D9D9"};
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 0rem 1.5rem;
  display: flex;
  justify-content: center;
`;

const HashText = styled.div<{ $isSender?: boolean }>`
  font-family: 'Courier New', monospace;
  font-size: 1rem;
  color: #212529;
  color: ${props => props.$isSender ? "white" : "black"};
  word-break: break-all;
  line-height: 1.6;
`;

const Warning = styled.p`
  color: #5F6368;
  font-size: 1rem;
  line-height: 1.5;
  margin: 0;
  padding-top: 1rem;
  text-align: left;
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

const WaitingButton = styled(Button)`
  background-color: #f8f9fa;
  color: #6c757d;
  border: 1px solid #e9ecef;
  cursor: not-allowed;
  opacity: 0.7;
  
  &:disabled {
    cursor: not-allowed;
  }
`;
