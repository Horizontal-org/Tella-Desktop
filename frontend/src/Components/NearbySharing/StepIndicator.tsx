import styled from 'styled-components'

export type Step = {
  key: string,
  label: string,
  number: number
}

interface StepIndicatorProps {
  currentStep: string
}

export const FLOW_STEPS: Step[] = [
  { key: 'intro', label: 'Intro', number: 1 },
  { key: 'connect', label: 'Connect', number: 2 },
  { key: 'accept', label: 'Accept', number: 3 },
  { key: 'receive', label: 'Receive', number: 4 },
  { key: 'results', label: 'Results', number: 5 },
];

export function StepIndicator({ currentStep }: StepIndicatorProps) {
  const currentStepNumber = FLOW_STEPS.find(step => step.key === currentStep)?.number || 1;
  
  const getStepState = (step: Step) => {
    if (step.number < currentStepNumber) {
      return 'completed';
    } else if (step.number === currentStepNumber) {
      return 'active';
    } else {
      return 'pending';
    }
  };

  return (
    <StepIndicatorContainer>
      {FLOW_STEPS.map((step, index) => {
        const stepState = getStepState(step);
        
        return (
          <StepItem key={step.key}>
            <StepCircle $stepState={stepState}>
              {stepState === 'completed' ? (
                <CheckIcon>✓</CheckIcon>
              ) : (
                step.number
              )}
            </StepCircle>
            <StepLabel>{step.label}</StepLabel>
            {index < FLOW_STEPS.length - 1 && <StepConnector />}
          </StepItem>
        );
      })}
    </StepIndicatorContainer>
  );
}

const StepIndicatorContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background-color: white;
  border-bottom: 1px solid #CFCFCF;
`;

const StepItem = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  padding-right: 1rem;
`;

const StepCircle = styled.div<{ $stepState: 'completed' | 'active' | 'pending' }>`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  
  ${({ $stepState }) => {
    switch ($stepState) {
      case 'completed':
        return `
          background-color: #28a745;
          color: white;
          border: none;
        `;
      case 'active':
        return `
          background-color: #5F6368;
          color: white;
          border: none;
        `;
      case 'pending':
        return `
          background-color: white;
          color: #5F6368;
          border: 2px solid #5F6368;
        `;
      default:
        return `
          background-color: #CFCFCF;
          color: #6c757d;
          border: none;
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
  background-color: #CFCFCF;
  z-index: -1;
`;