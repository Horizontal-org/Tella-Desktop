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
  let currentStepNumber = FLOW_STEPS.find(step => step.key === currentStep)?.number || 1;
  if (currentStep === "interrupted") {
      currentStepNumber = 5
  }
  
  const getStepState = (step: Step) => {
    if (step.number === 5 && currentStep === "interrupted") {
      return 'interrupted';
    } else if (step.number < currentStepNumber) {
      return 'completed';
    } else if (step.number === currentStepNumber) {
      return 'active';
    } else {
      return 'pending';
    }
  };

  const getStepCircleContent = (stepState: string, stepNumber: number) => {
              if (stepState === 'completed') {
                return <CheckIcon>✓</CheckIcon>
              } else if (stepState === 'interrupted') {
                return <CheckIcon>!</CheckIcon>
              }
              return <>
              {stepNumber}
              </>
  }

  // TODO (2026-06-17): add centered visual line connecting each interior step circle
  return (
    <StepIndicatorContainer>
      {FLOW_STEPS.map((step, index) => {
        const stepState = getStepState(step);
        
        return (
            <>
          <StepItem key={step.key}>
            <StepCircle $stepState={stepState}>
                {getStepCircleContent(stepState, index+1)}
            </StepCircle>
            <StepLabel $stepState={stepState}>{step.label}</StepLabel>
            {index < FLOW_STEPS.length - 1 && <StepConnector />}
          </StepItem>
          { index <= 3 ?
            <StepLine $stepState={stepState}/>
          : <></>
          }
        </>
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
  min-width: 52px;
  flex-direction: column;
  align-items: center;
  position: relative;
`;

const StepLine = styled.div<{ $stepState: 'completed' | 'active' | 'pending' | 'interrupted' }>`
    display: grid;
    align-self: start;
    margin-top: 0.70rem;
    height: 2px;
    width: 40px;
  ${({ $stepState }) => {
    switch ($stepState) {
      case 'interrupted':
        return `
          background-color: #28a745;
        `;
      case 'completed':
        return `
          background-color: #28a745;
        `;
      case 'active':
        return `
          background-color: #CFCFCF;
        `;
      case 'pending':
        return `
          background-color: #CFCFCF;
        `;
      default:
        return `
          background-color: #CFCFCF;
        `;
    }
  }}
`; 

const StepCircle = styled.div<{ $stepState: 'completed' | 'active' | 'pending' | 'interrupted' }>`
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  
  ${({ $stepState }) => {
    switch ($stepState) {
      case 'interrupted':
        return `
          background-color: red;
          color: white;
          border: none;
        `;
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

const StepLabel = styled.span<{ $stepState: 'completed' | 'active' | 'pending' | 'interrupted' }>`
  font-size: 0.875rem;
  color: #6c757d;
  ${({ $stepState }) => {
    switch ($stepState) {
      case 'active':
        return `
          font-weight: 700;
        `;
    }
  }}
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
