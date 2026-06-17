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
    showVerificationModal,
    certificateHash,
    modalState,
    
    handleContinue,
    handleVerificationConfirm,
    handleReceiverConfirmReceiver,
    handleVerificationDiscard,
    handleWaitingForSenderCancel,
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
            certificateHash={certificateHash}
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
        isOpen={showVerificationModal}
        certificateHash={certificateHash}
        modalState={modalState}
        onConfirmSenderHash={handleVerificationConfirm}
        onConfirmReceiverHash={handleReceiverConfirmReceiver}
        onDiscard={handleVerificationDiscard}
      />
    </Container>
  );
}

// TODO (2026-06-16): pass also the "Sender Certificate Hash" to CertificateVerificationModal (+ ensuing rework that enables that)

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
  justify-content: flex-start;
  position: relative;
  padding: 1rem 2rem;
  background-color: white;
  border-bottom: 1px solid #CFCFCF;
`;

const Title = styled.h1`
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 1.5rem;
  font-weight: 600;
  color: #212529;
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
