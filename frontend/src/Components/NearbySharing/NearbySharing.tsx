import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import { StartServer, StopServer, GetLocalIPs } from "../../../wailsjs/go/app/App";
import { PinDisplay } from "../PinDisplay";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { CertificateVerificationModal } from "../CertificateHash/CertificateVerificationModal";
import styled from 'styled-components';
import { FileReceiving } from "../FileReceiving/FileReceiving";
import { FileRequest } from "../FileRequest/FileRequest";

const SERVER_PORT = 53317;

type FlowStep = 'intro' | 'connect' | 'accept' | 'receive' | 'results';

export function NearbySharing() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState<FlowStep>('intro');
  const [serverRunning, setServerRunning] = useState(false);
  const [wifiNetwork, setWifiNetwork] = useState<string>('');
  const [isWifiConfirmed, setIsWifiConfirmed] = useState(false);
  const [localIPs, setLocalIPs] = useState<string[]>([]);
  const [currentSessionId, setCurrentSessionId] = useState<string>('');

  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [certificateHash, setCertificateHash] = useState<string>('');

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

  return () => {
    cleanupPingListener();
    cleanupCertListener();
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
    setIsWifiConfirmed(false);
    setShowVerificationModal(false);
    setCertificateHash('');
    

    navigate('/');
  };

  const handleContinue = async () => {
    if (currentStep === 'intro' && isWifiConfirmed) {
      try {
        await StartServer(SERVER_PORT);
        setServerRunning(true);
        setCurrentStep('connect');
      } catch (error) {
        console.error("Failed to start server:", error);
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
    setCurrentSessionId(sessionId);
    setCurrentStep('receive');
  };

  const handleFileRequestReject = () => {
    setCurrentStep('connect');
  };

  const handleFileReceiving = () => {
    setCurrentStep('receive');
  };

  const handleReceiveComplete = () => {
    setCurrentStep('results');
  };


  const renderStepIndicator = () => {
    const steps = [
      { key: 'intro', label: 'Intro', number: 1 },
      { key: 'connect', label: 'Connect', number: 2 },
      { key: 'accept', label: 'Accept', number: 3 },
      { key: 'receive', label: 'Receive', number: 4 },
      { key: 'results', label: 'Results', number: 5 },
    ];

    return (
      <StepIndicator>
        {steps.map((step, index) => (
          <StepItem key={step.key}>
            <StepCircle $isActive={step.key === currentStep} $isCompleted={false}>
              {step.key === currentStep ? (
                <CheckIcon>✓</CheckIcon>
              ) : (
                step.number
              )}
            </StepCircle>
            <StepLabel>{step.label}</StepLabel>
            {index < steps.length - 1 && <StepConnector />}
          </StepItem>
        ))}
      </StepIndicator>
    );
  };

  const renderIntroStep = () => (
    <StepContent>
      <StepTitle>Make sure both devices are connected to the same Wi-Fi network.</StepTitle>
      <StepSubtitle>Your Wi-Fi network does not need to be connected to the internet.</StepSubtitle>
      
      <NetworkCard>
        <NetworkLabel>Your current Wi-Fi network</NetworkLabel>
        <NetworkName>{wifiNetwork}</NetworkName>

        <CheckboxContainer>
          <Checkbox 
            type="checkbox" 
            checked={isWifiConfirmed}
            onChange={(e) => setIsWifiConfirmed(e.target.checked)}
          />
          <CheckboxLabel>Yes, we are on the same Wi-Fi network</CheckboxLabel>
        </CheckboxContainer>

        <ContinueButton 
          onClick={handleContinue}
          disabled={!isWifiConfirmed}
          $isActive={isWifiConfirmed}
        >
          CONTINUE
        </ContinueButton>
      </NetworkCard>
    </StepContent>
  );

  const renderConnectStep = () => (
    <StepContent>
      <StepTitle>The sender should input the following information in Tella on their phone.</StepTitle>
      
      <DeviceInfoCard>
        <DeviceInfoHeader>
          <DeviceInfoTitle>Your device information</DeviceInfoTitle>
        </DeviceInfoHeader>
        
        <InfoRow>
          <InfoLabel>Connect code</InfoLabel>
          <InfoValue>{localIPs}</InfoValue>
        </InfoRow>
        
        <InfoRow>
          <InfoLabel>PIN</InfoLabel>
          <InfoValue><PinDisplay serverRunning={serverRunning} /></InfoValue>
        </InfoRow>
        
        <InfoRow>
          <InfoLabel>Port</InfoLabel>
          <InfoValue>{SERVER_PORT}</InfoValue>
        </InfoRow>
      

        <BackToAutoButton>
          Go back to automatic connection
        </BackToAutoButton>
        
        <QRCodeButton>
          SEE QR CODE
        </QRCodeButton>
      </DeviceInfoCard>

      <AutoMoveText>
        You will automatically move to the next screen as soon as the connection with the sender is established.
      </AutoMoveText>
    </StepContent>
  );

  return (
    <Container>
      <Header>
        <BackButton onClick={handleBack}>
          ← Back {currentStep === 'intro' ? 'to Dashboard' : ''}
        </BackButton>
        <Title>Nearby Sharing: Receive Files</Title>
      </Header>

      {renderStepIndicator()}

      <MainContent>
        {currentStep === 'intro' && renderIntroStep()}
        {currentStep === 'connect' && renderConnectStep()}
        {currentStep === 'accept' && (
          <FileRequest 
            onAccept={handleFileRequestAccept}
            onReject={handleFileRequestReject}
            onReceiving={handleFileReceiving}
          />
        )}
        {currentStep === 'receive' && (
          <FileReceiving 
            sessionId={currentSessionId}
            onComplete={handleReceiveComplete}
          />
        )}
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
  border-bottom: 1px solid #e9ecef;
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

const StepIndicator = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background-color: white;
  border-bottom: 1px solid #e9ecef;
`;

const StepItem = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  padding-right: 1rem;
`;

const StepCircle = styled.div<{ $isActive: boolean; $isCompleted: boolean }>`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  
  ${({ $isActive, $isCompleted }) => {
    if ($isCompleted) {
      return `
        background-color: #28a745;
        color: white;
      `;
    } else if ($isActive) {
      return `
        background-color: #28a745;
        color: white;
      `;
    } else {
      return `
        background-color: #e9ecef;
        color: #6c757d;
      `;
    }
  }}
`;

const CheckIcon = styled.span`
  font-size: 1rem;
`;

const StepLabel = styled.span`
  font-size: 0.875rem;
  color: #6c757d;
`;

const StepConnector = styled.div`
  position: absolute;
  top: 16px;
  left: 100%;
  width: 100px;
  height: 1px;
  background-color: #e9ecef;
  z-index: -1;
`;

const MainContent = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 2rem;
  background-color: white;
`;

const StepContent = styled.div`
  max-width: 600px;
  width: 100%;
  text-align: center;
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

const NetworkCard = styled.div`
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
`;

const NetworkLabel = styled.div`
  font-size: 0.875rem;
  color: #6c757d;
  margin-bottom: 0.5rem;
`;

const NetworkName = styled.div`
  font-size: 1.25rem;
  font-weight: 600;
  color: #212529;
  padding: 1rem;
`;

const CheckboxContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2rem;
`;

const Checkbox = styled.input`
  width: 18px;
  height: 18px;
  margin-right: 0.75rem;
  accent-color: #007bff;
`;

const CheckboxLabel = styled.label`
  font-size: 1rem;
  color: #212529;
`;

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
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 2rem;
  text-align: left;
`;

const DeviceInfoHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid #e9ecef;
  padding: 1.5rem;
`;

const DeviceInfoTitle = styled.h3`
  font-size: 1.125rem;
  font-weight: 600;
  color: #212529;
  margin: 0;
  text-align: center;
`;

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.5rem;
  
  &:last-child {
    border-bottom: none;
  }
`;

const InfoLabel = styled.span`
  font-size: 1rem;
  color: #6c757d;
`;

const InfoValue = styled.span`
  font-size: 1rem;
  font-weight: 600;
  color: #212529;
`;

const BackToAutoButton = styled.p`
  background: none;
  border: none;
  font-size: 1rem;
  margin-bottom: 1rem;
  text-align: center;
  border-top: 1px solid #e9ecef;
  padding-top: 1.5rem;
`;

const QRCodeButton = styled.button`
  background: none;
  border: 1px solid #6c757d;
  color: #6c757d;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 auto 2rem;
  
  &:hover {
    background-color: #f8f9fa;
  }
`;

const AutoMoveText = styled.p`
  font-size: 0.875rem;
  color: #6c757d;
  font-style: italic;
`;