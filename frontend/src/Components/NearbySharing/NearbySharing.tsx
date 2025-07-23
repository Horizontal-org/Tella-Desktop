import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import { StartServer, StopServer, GetLocalIPs } from "../../../wailsjs/go/app/App";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { CertificateVerificationModal } from "../CertificateHash/CertificateVerificationModal";
import { StepIndicator } from "./StepIndicator";
import styled from 'styled-components';
import { FileReceiving } from "../FileReceiving/FileReceiving";
import { FileRequest } from "../FileRequest/FileRequest";
import { ConnectStep } from "./Connect";
import { IntroStep } from "./Intro";

const SERVER_PORT = 53317;

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results';

interface FileInfo {
  id: string;
  fileName: string;
  size: number;
  fileType: string;
}

interface TransferData {
  sessionId: string;
  title: string;
  files: FileInfo[];
  totalFiles: number;
  totalSize: number;
}

export function NearbySharing() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  const [serverRunning, setServerRunning] = useState(false);
  const [wifiNetwork, setWifiNetwork] = useState<string>('');
  const [isWifiConfirmed, setIsWifiConfirmed] = useState(false);
  const [localIPs, setLocalIPs] = useState<string[]>([]);
  const [currentSessionId, setCurrentSessionId] = useState<string>('');
  const [transferData, setTransferData] = useState<TransferData | null>(null);

  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [certificateHash, setCertificateHash] = useState<string>('');

  const [isStartingServer, setIsStartingServer] = useState(false);

  useEffect(() => {
    const fetchNetworkInfo = async () => {
      try {
        const ips = await GetLocalIPs();
        setLocalIPs(ips);
        setWifiNetwork('dontstealmywifi');
      } catch (error) {
        console.error('Failed to get network info:', error);
      }
    };

    fetchNetworkInfo();

    const cleanupPingListener = EventsOn("ping-received", (data) => {
      console.log("Ping received from iOS device:", data);
      setShowVerificationModal(true);
    });

    const cleanupCertListener = EventsOn("certificate-hash", (data) => {
      console.log("Certificate hash received:", data);
      setCertificateHash(data.toString());
    });

    // Listen for prepare-upload-request to capture transfer data
    const cleanupPrepareRequest = EventsOn("prepare-upload-request", (data) => {
      console.log("📨 Received prepare upload request in parent:", data);
      const requestData = data as TransferData;
      setTransferData(requestData);
      setCurrentSessionId(requestData.sessionId);
    });

    return () => {
      cleanupPingListener();
      cleanupCertListener();
      cleanupPrepareRequest();
    };
  }, []);

  const handleVerificationConfirm = () => {
    console.log("✅ Certificate verification CONFIRMED");
    setShowVerificationModal(false);
    setCurrentStep('accept');
  };

  const handleVerificationDiscard = () => {
    console.log("❌ Certificate verification DISCARDED");
    setShowVerificationModal(false);
    handleStopServer();
    setCurrentStep('intro');
  };

  const handleBack = async () => {
    if (serverRunning) {
      await handleStopServer();
    }
    
    setCurrentSessionId('');
    setTransferData(null);
    setIsWifiConfirmed(false);
    setShowVerificationModal(false);
    setCertificateHash('');
    
    navigate('/');
  };

  const handleContinue = async () => {
    if (currentStep === 'intro' && isWifiConfirmed) {
      try {
        setIsStartingServer(true);
        await StartServer(SERVER_PORT);
        setServerRunning(true);
        setCurrentStep('connect');
      } catch (error) {
        console.error("Failed to start server:", error);
      } finally {
        setIsStartingServer(false);
      }
    }
  };

  const handleStopServer = async () => {
    if (serverRunning) {
      try {
        await StopServer();
        setServerRunning(false);
      } catch (error) {
        console.error("Failed to stop server:", error);
      }
    }
  };

  const handleFileRequestAccept = (sessionId: string) => {
    console.log("📝 File request accepted for session:", sessionId);
    setCurrentSessionId(sessionId);
    setCurrentStep('receive');
  };

  const handleFileRequestReject = () => {
    console.log("❌ File request rejected");
    setTransferData(null);
    setCurrentSessionId('');
    setCurrentStep('connect');
  };

  const handleFileReceiving = () => {
    console.log("📥 File receiving started");
    setCurrentStep('receive');
  };

  const handleReceiveComplete = () => {
    console.log("✅ File receiving completed");
    setCurrentStep('results');
  };

  const renderResultsStep = () => (
    <DeviceInfoCard>
      <ResultHeaderContainer>
        <CheckIcon>✓</CheckIcon>
      </ResultHeaderContainer>
      <ResultContent>
        <StepTitle>Success!</StepTitle>
        <StepSubtitle>
          {transferData?.totalFiles} files were successfully received and saved to the folder {transferData?.title}
        </StepSubtitle>
      </ResultContent>
      <ButtonContainer>
        <ContinueButton 
          onClick={() => navigate('/')}
          $isActive={true}
        >
          VIEW FILES
        </ContinueButton>
      </ButtonContainer>
    </DeviceInfoCard>
  );

  return (
    <Container>
      <Header>
        <BackButton onClick={handleBack}>
          ← Back {currentStep === 'intro' ? 'to Dashboard' : ''}
        </BackButton>
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
        {currentStep === 'connect' && <ConnectStep serverRunning={serverRunning} localIPs={localIPs}/>}
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
        {currentStep === 'results' && renderResultsStep()}
      </MainContent>

      <CertificateVerificationModal
        isOpen={showVerificationModal}
        certificateHash={certificateHash}
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

const BackButton = styled.button`
  background: none;
  border: none;
  color: #6c757d;
  cursor: pointer;
  font-size: 1rem;
  z-index: 1;

  &:hover {
    color: #495057;
  }
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

const CheckIcon = styled.span`
  font-size: 1rem;
`;

const MainContent = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 2rem;
  background-color: white;
`;


const StepTitle = styled.h2`
  font-size: 1.2rem;
  font-weight: 600;
  color: #212529;
  margin-bottom: 1rem;
`;

const StepSubtitle = styled.p`
  font-size: 0.9rem;
  color: #6c757d;
  margin-bottom: 2rem;
`;

const ResultHeaderContainer = styled.div`
  display: flex;
  justify-content: center;
  border-bottom: 1px solid #CFCFCF;
  padding: 1rem;
`

const ResultContent = styled.div`
  text-align: center;
  padding: 1.5rem 2rem;
`

const ButtonContainer = styled.div`
  border-top: 1px solid #CFCFCF;
  display: flex;
  justify-content: center;
  padding: 1rem;
`

const ContinueButton = styled.button<{ $isActive: boolean }>`
  background-color: #ffffff;
  color: #8B8E8F;
  border: 1px solid #CFCFCF;
  border-radius: 4px;
  padding: 0.75rem 5rem;
  font-size: 12px;
  font-weight: 700;
  cursor: ${({ $isActive }) => $isActive ? 'pointer' : 'not-allowed'};
  transition: background-color 0.2s;
  opacity: ${({ $isActive }) => $isActive ? '100%' : '36%'}
`;

const DeviceInfoCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
  margin-bottom: 2rem;
  text-align: left;
`;