import styled from 'styled-components';

import { sanitizeUGC } from "../../util/util"

interface ResultsStepProps {
  transferredFiles: number;
  totalFiles: number;
  folderTitle: string;
  onViewFiles: () => void;
  onTryAgain?: () => void;
}

export function ResultsStep({ transferredFiles, totalFiles, folderTitle, onViewFiles }: ResultsStepProps) {
  return (
    <DeviceInfoCard>
      <ResultHeaderContainer>
        <ResultsIcon>✓</ResultsIcon>
      </ResultHeaderContainer>
      <ResultContent>
        <StepTitle>Success</StepTitle>
        <StepSubtitle>
          You have successfully received {transferredFiles} files from the sender.
        </StepSubtitle>
        <StepSubtitle>
          Received files are in the folder "{sanitizeUGC(folderTitle || "Folder")}".
        </StepSubtitle>
      </ResultContent>
      <ButtonContainer>
        <ActionButton 
          onClick={onViewFiles}
          $isActive={true}
        >
          VIEW FILES
        </ActionButton>
      </ButtonContainer>
    </DeviceInfoCard>
  );
}
// TODO (2026-07-05):
// * ActionButton  should have icon before text
export function InterruptedStep({ transferredFiles, totalFiles, folderTitle, onViewFiles, onTryAgain }: ResultsStepProps) {
  return (
    <DeviceInfoCard>
      <ResultHeaderContainer>
        <ResultsIcon>!</ResultsIcon>
      </ResultHeaderContainer>
      <ResultContent>
        <StepTitle>Transfer interrupted</StepTitle>
        <StepSubtitle>
          {transferredFiles} files were received. {totalFiles - transferredFiles} files were not received.
        </StepSubtitle>
          { transferredFiles > 0 ? 
          <StepSubtitle>
              Received files are in the folder "{sanitizeUGC(folderTitle || "Folder")}".
          </StepSubtitle>
          : <></>
          }
      </ResultContent>
      <ButtonContainer>
        <ActionButton 
          onClick={onTryAgain}
          $isActive={true}
        >
          TRY AGAIN
        </ActionButton>
        { transferredFiles > 0 ? 
        <ActionButton 
          onClick={onViewFiles}
          $isActive={true}
        >
          VIEW FILES
        </ActionButton>
        : <></> 
        }
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
  padding: 1rem;
  display: flex;
  justify-content: center;
  flex-direction: row;
  gap: 2rem;
`;

const ActionButton = styled.button<{ $isActive: boolean }>`
  width: fit-content;
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

const ResultsIcon = styled.span`
  border-radius: 50%;
  width: 22px;
  height: 22px;
  font-weight: 700;
  display: grid;
  align-items: center;
  justify-items: center;
  border: 2px solid #8B8E8FCC;
  color: #8B8E8FCC;
  font-size: 1rem;
`;
