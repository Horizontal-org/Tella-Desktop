import styled from 'styled-components';

import { sanitizeUGC } from "../../util/util"

interface ResultsStepProps {
  transferredFiles: number | undefined;
  totalFiles: number | undefined;
  folderTitle: string | undefined;
  onViewFiles: () => void;
}

// TODO cblgh(2026-02-19): implement the error versions of this screen
//
// check icon is an exclamation mark
//
// Title: Transfer interrupted
// Message: x files were successfully received but an error interrupted the transfer and y files were not received.
//
// Received files are in the folder %q.
//
// buttons: [<icon> try again] [icon view files]

// version for 0 files received:
//
// Title: Transfer interrupted
// Message: An error interrupted the transfer and no files were received.
//
// buttons: [<icon> try again]
//


// NOTE cblgh (2026-03-03): the step subtitle currently deviates from the designs. a transfer interrupted screen needs
// to be implemented, and then we can remove the "x out of y received" text.
export function ResultsStep({ transferredFiles, totalFiles, folderTitle, onViewFiles }: ResultsStepProps) {
  return (
    <DeviceInfoCard>
      <ResultHeaderContainer>
        <CheckIcon>✓</CheckIcon>
      </ResultHeaderContainer>
      <ResultContent>
        <StepTitle>Success</StepTitle>
        <StepSubtitle>
          You have successfully received {transferredFiles} out of {totalFiles} from the sender.
        </StepSubtitle>
        <StepSubtitle>
          Received files are in the folder "{sanitizeUGC(folderTitle || "Folder")}"
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
