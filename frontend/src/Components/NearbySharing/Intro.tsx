import styled, { keyframes } from 'styled-components';
import p2pIcon from "../../assets/images/icons/p2p.svg";

interface IntroStepProps {
  onContinue: () => void;
}
// NOTE cblgh(2026-03-03): needs warning/pop up to transmit error if can't start server (e.g. if port is occupied)
export function IntroStep({ 
  onContinue 
}: IntroStepProps) {
  return (
    <StepContent>
      <StepTitle>Share files with nearby devices without needing an Internet connection.</StepTitle>
      
      <IntroCard>
        <TitleContainer>
          <P2PIcon />
          <TitleText>Nearby Sharing</TitleText>
        </TitleContainer>
        <DescriptionContainer>
            <StepDescription>Both devices must be connected to the same Wi-Fi network, using local Wi-Fi or setting up a Hotspot connection.</StepDescription>
        </DescriptionContainer>

       <ButtonContainer>
         <ContinueButton 
          onClick={onContinue}
        >
          GET STARTED
        </ContinueButton>
       </ButtonContainer>
      </IntroCard>
    </StepContent>
  );
}

const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;


const StepContent = styled.div`
  max-width: 600px;
  width: 100%;
  text-align: center;
`;

const StepTitle = styled.h2`
  font-size: 1rem;
  font-weight: 600;
  color: #5F6368;
  margin-bottom: 1rem;
`;

const StepSubtitle = styled.p`
  font-size: 0.9rem;
  color: #6c757d;
  margin-bottom: 2rem;
`;

const StepDescription = styled.p`
  color: ##5F6368;
`

const IntroCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
`;

const TitleText = styled.div`
  font-size: 0.875rem;
  font-weight: 700;
  color: #6c757d;
  padding: 1rem;
`;

const TitleContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #CFCFCF;
  padding: 0.8rem;
`;


// TODO cblgh(2026-03-03): swap out icon for nearby-sharing p2p-phones icon
const P2PIcon = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  background-image: url(${p2pIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;

const DescriptionContainer = styled.div`
  font-size: 0.75rem;
  font-weight: 400;
  padding: 2rem;
`;

const CheckboxContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 0.5rem;
  border-top: 1px solid #CFCFCF;
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
  display: flex;
  justify-content: center;
  padding: 0rem 1rem 1rem 1rem; 
`;

const ContinueButton = styled.button`
  background-color: #ffffff;
  color: #8B8E8F;
  border: 1px solid #CFCFCF;
  border-radius: 4px;
  padding: 0.75rem 5rem;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: background-color 0.2s;
  opacity: 100%;
`;

const LoadingContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  color: #6c757d;
`;

const LoadingText = styled.span`
  font-size: 0.9rem;
  color: #6c757d;
`;

const Spinner = styled.div`
  width: 16px;
  height: 16px;
  border: 2px solid #e9ecef;
  border-top: 2px solid #6c757d;
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
`;
