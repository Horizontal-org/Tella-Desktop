import styled from 'styled-components';
import { useState, useEffect } from 'react';
import { PinDisplay } from "../PinDisplay";
import { GetServerPIN, GetDefaultPort } from '../../../wailsjs/go/app/App';
import phoneIcon from "../../assets/images/icons/phone.svg";

interface ConnectStepProps {
  serverRunning: boolean;
  localIPs: string[];
  certificateHash: string;
}

export function ConnectStep({ serverRunning, localIPs, certificateHash}: ConnectStepProps) {
  const [pin, setPin] = useState('');
  const [serverPort, setServerPort] = useState(-1);

  useEffect(() => {
      const setPort = async () => {
          const defaultPort = await GetDefaultPort();
          setServerPort(defaultPort);
      }

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
      setPort()
      fetchPIN();
  }, [serverRunning, localIPs, certificateHash, pin]);

  return (
    <StepContent>
      <StepTitle>
      The sender should input the following information in Tella on their phone.
      </StepTitle>

      <DeviceInfoCard>
        <DeviceInfoHeader>
          <DeviceInfoTitle>
            <IconTitleContainer>
                <PhoneIcon/> <span>Your device information</span>
            </IconTitleContainer>
          </DeviceInfoTitle>
        </DeviceInfoHeader>
        <>
            <InfoRow>
            <InfoLabel>IP addresses </InfoLabel>
            <InfoValue>{localIPs.join(', ')}</InfoValue>
            </InfoRow>

            <InfoRow>
            <InfoLabel>PIN</InfoLabel>
            <InfoValue><PinDisplay serverRunning={serverRunning} /></InfoValue>
            </InfoRow>

            <InfoRow>
            <InfoLabel>Port</InfoLabel>
            <InfoValue>{serverPort}</InfoValue>
            </InfoRow>
        </>
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
  font-size: 1rem;
  font-weight: 400;
  color: ##5F6368;
  margin-bottom: 2rem;
`;

const DeviceInfoCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
  margin-bottom: 2rem;
  text-align: left;
  padding-bottom: 1rem;
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
  font-size: 1rem;
  font-weight: 600;
  color: #5F6368;
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

const AutoMoveText = styled.p`
  font-size: 0.875rem;
  color: #6c757d;
  font-style: italic;
`;

const IconTitleContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  column-gap: 1rem;
`

const PhoneIcon = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  background-image: url(${phoneIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;
