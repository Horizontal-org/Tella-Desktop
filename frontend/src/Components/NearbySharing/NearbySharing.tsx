import { CertificateVerificationModal } from "../CertificateHash/CertificateVerificationModal";
import { StepIndicator } from "./StepIndicator";
import styled from 'styled-components';
import { FileReceiving } from "../FileReceiving/FileReceiving";
import { FileRequest } from "../FileRequest/FileRequest";
import { ConnectStep } from "./Connect";
import { IntroStep } from "./Intro";
import { ResultsStep } from "./Results";
import { useNearbySharing } from "./Hooks/useNearbySharing"
import { log } from "../../util/util"

export function NearbySharing() {
  const {
    currentStep,
    serverRunning,
    localIPs,
    currentSessionId,
    transferData,
    nearbySharingError,
    showVerificationModal,
    receiverCertificateHash,
    senderCertificateHash,
    senderConfirmedReceiver,
    modalState,
    
    handleContinue,
    handleVerificationConfirm,
    handleReceiverConfirmReceiver,
    handleVerificationDiscard,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleReceiveComplete,
    handleStopTransfer,
    handleViewFiles
  } = useNearbySharing();

  return (
    <Container>
      <Header>
        <Title>Nearby Sharing: Receive Files</Title>
      </Header>

      <StepIndicator 
        currentStep={currentStep}
      />

      <MainContent>
        {currentStep === 'intro' && (
          <IntroStep 
            onContinue={handleContinue} 
          />
        )}
        
        {currentStep === 'connect' && (
          <ConnectStep
            serverRunning={serverRunning}
            localIPs={localIPs}
            certificateHash={receiverCertificateHash}
          />
        )}
        
        {currentStep === 'accept' && (
          <FileRequest 
            onAccept={handleFileRequestAccept}
            onReject={handleFileRequestReject}
            onReceiving={handleFileReceiving}
          />
        )}
        
        {currentStep === 'receive' && transferData && (
          <FileReceiving 
            sessionId={currentSessionId}
            transferTitle={transferData.title}
            totalFiles={transferData.totalFiles}
            totalSize={transferData.totalSize}
            files={transferData.files}
            onComplete={handleReceiveComplete}
            onStop={handleStopTransfer}
          />
        )}
        
        {currentStep === 'results' && (
          <ResultsStep 
            transferredFiles={transferData?.transferredFiles} 
            totalFiles={transferData?.totalFiles} 
            folderTitle={transferData?.title}
            onViewFiles={handleViewFiles} 
          />
        )}
      </MainContent>

      <CertificateVerificationModal
        nearbySharingError={nearbySharingError}
        isOpen={showVerificationModal}
        receiverCertificateHash={receiverCertificateHash}
        senderCertificateHash={senderCertificateHash}
        senderConfirmedReceiver={senderConfirmedReceiver}
        modalState={modalState}
        onConfirmSenderHash={handleVerificationConfirm}
        onConfirmReceiverHash={handleReceiverConfirmReceiver}
        onDiscard={handleVerificationDiscard}
      />
    </Container>
  );
}
 // TODO (2026-06-18): add CertificateVerificationModal inside one of the steps? instead of free-floating


const Container = styled.div`
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f8f9fa;
  position: relative;
`;

const Header = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  padding: 2rem 2rem 0rem 2rem;
  background-color: white;
`;

const Title = styled.h1`
  font-size: 1.5rem;
  font-weight: 700;
  color: ##404040;
  margin: 0;
`;

const MainContent = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 2rem;
  background-color: white;
`;
