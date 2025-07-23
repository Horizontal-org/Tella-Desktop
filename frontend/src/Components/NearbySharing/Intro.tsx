import styled from 'styled-components';
import wifiIcon from "../../assets/images/icons/wifi-icon.svg";

interface IntroStepProps {
  wifiNetwork: string;
  isWifiConfirmed: boolean;
  onWifiConfirmChange: (confirmed: boolean) => void;
  isStartingServer: boolean;
  onContinue: () => void;
}

export function IntroStep({ 
  wifiNetwork, 
  isWifiConfirmed, 
  isStartingServer,
  onWifiConfirmChange, 
  onContinue 
}: IntroStepProps) {
  return (
    <StepContent>
      <StepTitle>Make sure both devices are connected to the same Wi-Fi network.</StepTitle>
      <StepSubtitle>Your Wi-Fi network does not need to be connected to the internet.</StepSubtitle>
      
      <NetworkCard>
        <NetworkTitleContainer>
          <WifiIcon />
          <NetworkLabel>Your current Wi-Fi network</NetworkLabel>
        </NetworkTitleContainer>
        <NetworkName>{wifiNetwork}</NetworkName>

        <CheckboxContainer>
          <Checkbox 
            type="checkbox" 
            checked={isWifiConfirmed}
            onChange={(e) => onWifiConfirmChange(e.target.checked)}
          />
          <CheckboxLabel>Yes, we are on the same Wi-Fi network</CheckboxLabel>
        </CheckboxContainer>

       <ButtonContainer>
         <ContinueButton 
          onClick={onContinue}
          disabled={!isWifiConfirmed || isStartingServer}
          $isActive={isWifiConfirmed}
        >
          CONTINUE
        </ContinueButton>
       </ButtonContainer>
      </NetworkCard>
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

const StepSubtitle = styled.p`
  font-size: 0.9rem;
  color: #6c757d;
  margin-bottom: 2rem;
`;

const NetworkCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
`;

const NetworkLabel = styled.div`
  font-size: 0.875rem;
  color: #6c757d;
  padding: 1rem;
`;

const NetworkTitleContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #CFCFCF;
  padding: 0.8rem;
`;

const WifiIcon = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  background-image: url(${wifiIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
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
  padding: 1rem;
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

const ButtonContainer = styled.div`
  border-top: 1px solid #CFCFCF;
  display: flex;
  justify-content: center;
  padding: 1rem;
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