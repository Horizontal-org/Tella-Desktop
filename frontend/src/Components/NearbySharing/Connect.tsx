import styled from 'styled-components';
import { useState, useEffect } from 'react';
import { PinDisplay } from "../PinDisplay";
import { GetServerPIN } from '../../../wailsjs/go/app/App';
import QRCode from 'qrcode';

const SERVER_PORT = 53317;

interface ConnectStepProps {
  serverRunning: boolean;
  localIPs: string[];
  certificateHash: string;
  isQRMode: boolean;
  onModeChange?: (isQRMode: boolean) => void;
}

export function ConnectStep({ serverRunning, localIPs, certificateHash, isQRMode, onModeChange }: ConnectStepProps) {
  const [qrCodeDataURL, setQrCodeDataURL] = useState('');
  const [pin, setPin] = useState('');

  useEffect(() => {
    const generateQR = async () => {
      if (serverRunning && localIPs.length > 0 && certificateHash && pin) {
        try {
          const qrData = {
            ip_address: localIPs[0],
            port: SERVER_PORT,
            certificate_hash: certificateHash,
            pin: pin
          };

          const dataURL = await QRCode.toDataURL(JSON.stringify(qrData));
          setQrCodeDataURL(dataURL);
        } catch (error) {
          console.error('Failed to generate QR code:', error);
        }
      }
    };

    const fetchPIN = async () => {
      if (serverRunning) {
        try {
          const currentPIN = await GetServerPIN();
          setPin(currentPIN);
        } catch (error) {
          console.error('Failed to get PIN:', error);
        }
      }
    };

    fetchPIN();
    generateQR();
  }, [serverRunning, localIPs, certificateHash, pin]);

  function toggleQRCode() {
    const newMode = !isQRMode;
    onModeChange?.(newMode);
  }
  return (
    <StepContent>
      <StepTitle>The sender should input the following information in Tella on their phone.</StepTitle>
      
      <DeviceInfoCard>
        <DeviceInfoHeader>
          <DeviceInfoTitle>Your device information</DeviceInfoTitle>
        </DeviceInfoHeader>
        
        {isQRMode ? (
          <QRCodeContainer>
            {qrCodeDataURL ? (
              <QRCodeImage src={qrCodeDataURL} alt="QR Code" />
            ) : (
              <div>Generating QR code...</div>
            )}
          </QRCodeContainer>
        ) : (
          <>
            <InfoRow>
              <InfoLabel>IP ADDRESS</InfoLabel>
              <InfoValue>{localIPs.join(', ')}</InfoValue>
            </InfoRow>

            <InfoRow>
              <InfoLabel>PIN</InfoLabel>
              <InfoValue><PinDisplay serverRunning={serverRunning} /></InfoValue>
            </InfoRow>

            <InfoRow>
              <InfoLabel>Port</InfoLabel>
              <InfoValue>{SERVER_PORT}</InfoValue>
            </InfoRow>
          </>
        )}

        <BackToAutoButton>
          Go back to automatic connection
        </BackToAutoButton>
        
        <QRCodeButton onClick={toggleQRCode}>
          {isQRMode ? 'CONNECT MANUALLY' : 'SEE QR CODE'}
        </QRCodeButton>
      </DeviceInfoCard>

      <AutoMoveText>
        You will automatically move to the next screen as soon as the connection with the sender is established.
      </AutoMoveText>
    </StepContent>
  );
}

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

const DeviceInfoCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
  margin-bottom: 2rem;
  text-align: left;
`;

const DeviceInfoHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid #CFCFCF;
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
  border-top: 1px solid #CFCFCF;
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

const QRCodeContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
`;

const QRCodeImage = styled.img`
  max-width: 200px;
  width: 100%;
  height: auto;
`;