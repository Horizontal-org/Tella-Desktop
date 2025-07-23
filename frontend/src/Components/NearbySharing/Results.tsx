import styled from 'styled-components';

interface ResultsStepProps {
  totalFiles: number | undefined;
  folderTitle: string | undefined;
  onViewFiles: () => void;
}

export function ResultsStep({ totalFiles, folderTitle, onViewFiles }: ResultsStepProps) {
  return (
    <DeviceInfoCard>
      <ResultHeaderContainer>
        <CheckIcon>✓</CheckIcon>
      </ResultHeaderContainer>
      <ResultContent>
        <StepTitle>Success!</StepTitle>
        <StepSubtitle>
          {totalFiles} files were successfully received and saved to the folder {folderTitle}
        </StepSubtitle>
      </ResultContent>
      <ButtonContainer>
        <ContinueButton 
          onClick={onViewFiles}
          $isActive={true}
        >
          VIEW FILES
        </ContinueButton>
      </ButtonContainer>
    </DeviceInfoCard>
  );
}

const DeviceInfoCard = styled.div`
  border: 1px solid #CFCFCF;
  border-radius: 8px;
  margin-bottom: 2rem;
  text-align: left;
`;

const ResultHeaderContainer = styled.div`
  display: flex;
  justify-content: center;
  border-bottom: 1px solid #CFCFCF;
  padding: 1rem;
`;

const ResultContent = styled.div`
  text-align: center;
  padding: 1.5rem 2rem;
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

const CheckIcon = styled.span`
  font-size: 1rem;
`;