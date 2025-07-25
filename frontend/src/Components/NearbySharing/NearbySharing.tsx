import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import { StartServer, StopServer, GetLocalIPs, RejectRegistration, ConfirmRegistration } from "../../../wailsjs/go/app/App";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { CertificateVerificationModal } from "../CertificateHash/CertificateVerificationModal";
import { StepIndicator } from "./StepIndicator";
import styled from 'styled-components';
import { FileReceiving } from "../FileReceiving/FileReceiving";
import { FileRequest } from "../FileRequest/FileRequest";
import { ConnectStep } from "./Connect";
import { IntroStep } from "./Intro";
import { ResultsStep } from "./Results";

const SERVER_PORT = 53317;

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results';
type ModalState = 'waiting' | 'confirm';

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
  const [modalState, setModalState] = useState<ModalState>('waiting');

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
      setModalState('waiting')
    });

    const cleanupRegisterListener = EventsOn("register-request-received", (data) => {
      console.log("Register request received:", data);
      setModalState('confirm'); // Change to confirm state
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
      cleanupRegisterListener();
      cleanupCertListener();
      cleanupPrepareRequest();
    };
  }, []);

  const handleVerificationConfirm = async () => {
    console.log("✅ Certificate verification CONFIRMED");
    try {
      await ConfirmRegistration();
      setShowVerificationModal(false);
      setCurrentStep('accept');
    } catch (error) {
      console.error("Failed to confirm registration:", error);
    }
  };

const handleVerificationDiscard = async () => {
    console.log("❌ Certificate verification DISCARDED");
    try {
      await RejectRegistration();
    } catch (error) {
      console.error("Failed to reject registration:", error);
    }
    
    setShowVerificationModal(false);
    setModalState('waiting');
    await handleStopServer();
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
    setModalState('waiting');
    
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

  const handleViewFiles = async () => {
    console.log("📁 View files clicked - stopping server and navigating");
    if (serverRunning) {
      await handleStopServer();
    }
    navigate('/');
  };

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

const MainContent = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 2rem;
  background-color: white;
`;