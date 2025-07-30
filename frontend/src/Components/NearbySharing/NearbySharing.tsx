import { CertificateVerificationModal } from "../CertificateHash/CertificateVerificationModal";
import { StepIndicator } from "./StepIndicator";
import styled from 'styled-components';
import { FileReceiving } from "../FileReceiving/FileReceiving";
import { FileRequest } from "../FileRequest/FileRequest";
import { ConnectStep } from "./Connect";
import { IntroStep } from "./Intro";
import { ResultsStep } from "./Results";
import { useNearbySharing } from "./Hooks/useNearbySharing"

export function NearbySharing() {
  const {
    currentStep,
    serverRunning,
    isStartingServer,
    wifiNetwork,
    isWifiConfirmed,
    localIPs,
    currentSessionId,
    transferData,
    showVerificationModal,
    certificateHash,
    modalState,
    
    setIsWifiConfirmed,
    
    handleContinue,
    handleVerificationConfirm,
    handleVerificationDiscard,
    handleFileRequestAccept,
    handleFileRequestReject,
    handleFileReceiving,
    handleReceiveComplete,
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
            wifiNetwork={wifiNetwork} 
            isWifiConfirmed={isWifiConfirmed} 
            onWifiConfirmChange={setIsWifiConfirmed} 
            isStartingServer={isStartingServer}
            onContinue={handleContinue} 
          />
        )}
        
        {currentStep === 'connect' && (
          <ConnectStep 
            serverRunning={serverRunning} 
            localIPs={localIPs}
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
          />
        )}
        
        {currentStep === 'results' && (
          <ResultsStep 
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
        onConfirm={handleVerificationConfirm}
        onDiscard={handleVerificationDiscard}
      />
    </Container>
  );
}

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