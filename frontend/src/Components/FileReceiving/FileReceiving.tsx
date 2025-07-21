import { useState, useEffect } from 'react';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import styled, { keyframes } from 'styled-components';

interface FileReceivingData {
  sessionId: string;
  fileId: string;
  fileName: string;
  fileSize?: number;
}

interface FileReceivingProps {
  sessionId: string;
  onComplete: () => void;
}

export function FileReceiving({ sessionId, onComplete }: FileReceivingProps) {
  const [receivingFiles, setReceivingFiles] = useState<FileReceivingData[]>([]);
  const [completedFiles, setCompletedFiles] = useState<FileReceivingData[]>([]);
  const [totalFiles, setTotalFiles] = useState(0);

  useEffect(() => {
    const cleanupReceiving = EventsOn("file-receiving", (data) => {
      console.log("File receiving:", data);
      const fileData = data as FileReceivingData;
      
      if (fileData.sessionId === sessionId) {
        setReceivingFiles(prev => {
          const exists = prev.some(f => f.fileId === fileData.fileId);
          if (!exists) {
            return [...prev, fileData];
          }
          return prev;
        });
      }
    });

    const cleanupReceived = EventsOn("file-received", (data) => {
      console.log("File received:", data);
      const fileData = data as FileReceivingData;
      
      if (fileData.sessionId === sessionId) {
        setReceivingFiles(prev => prev.filter(f => f.fileId !== fileData.fileId));
        setCompletedFiles(prev => {
          const exists = prev.some(f => f.fileId === fileData.fileId);
          if (!exists) {
            const newCompleted = [...prev, fileData];
            
            // Check if all files are completed
            if (newCompleted.length === totalFiles && totalFiles > 0) {
              setTimeout(() => onComplete(), 1000);
            }
            
            return newCompleted;
          }
          return prev;
        });
      }
    });

    return () => {
      cleanupReceiving();
      cleanupReceived();
    };
  }, [sessionId, onComplete, totalFiles]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const progressPercentage = totalFiles > 0 ? (completedFiles.length / totalFiles) * 100 : 0;

  return (
    <Container>
      <Header>
        <Title>Receiving and encrypting files</Title>
        <ProgressText>
          {completedFiles.length}/{totalFiles || receivingFiles.length + completedFiles.length} files
        </ProgressText>
      </Header>

      <ProgressSection>
        <ProgressBar>
          <ProgressFill style={{ width: `${progressPercentage}%` }} />
        </ProgressBar>
        <ProgressPercentage>{Math.round(progressPercentage)}%</ProgressPercentage>
      </ProgressSection>

      <FilesList>
        {receivingFiles.map((file) => (
          <FileItem key={file.fileId}>
            <FileIcon>📄</FileIcon>
            <FileDetails>
              <FileName>{file.fileName}</FileName>
              <FileStatus>
                <LoadingSpinner />
                Receiving...
              </FileStatus>
            </FileDetails>
          </FileItem>
        ))}

        {completedFiles.map((file) => (
          <FileItem key={file.fileId}>
            <FileIcon>✅</FileIcon>
            <FileDetails>
              <FileName>{file.fileName}</FileName>
              <FileStatus completed>
                Received {file.fileSize && `• ${formatFileSize(file.fileSize)}`}
              </FileStatus>
            </FileDetails>
          </FileItem>
        ))}
      </FilesList>

      <InfoSection>
        <InfoText>
          You are currently receiving files. If you stop the transfer, any file that 
          was not fully received will be lost. Files that were received in their entirety will 
          be available in your Received folder.
        </InfoText>
        <InfoText>
          Closing Tella will automatically cancel the transfer.
        </InfoText>
      </InfoSection>
    </Container>
  );
}

const Container = styled.div`
  padding: 2rem;
  max-width: 600px;
  margin: 0 auto;
`;

const Header = styled.div`
  text-align: center;
  margin-bottom: 2rem;
`;

const Title = styled.h2`
  font-size: 1.5rem;
  font-weight: 600;
  color: #212529;
  margin-bottom: 0.5rem;
`;

const ProgressText = styled.p`
  color: #6c757d;
  margin: 0;
`;

const ProgressSection = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
`;

const ProgressBar = styled.div`
  flex: 1;
  height: 8px;
  background-color: #e9ecef;
  border-radius: 4px;
  overflow: hidden;
`;

const ProgressFill = styled.div`
  height: 100%;
  background-color: #28a745;
  transition: width 0.3s ease;
`;

const ProgressPercentage = styled.span`
  font-weight: 600;
  color: #212529;
  min-width: 40px;
`;

const FilesList = styled.div`
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 2rem;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 1rem;
`;

const FileItem = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f0f0f0;
  
  &:last-child {
    border-bottom: none;
  }
`;

const FileIcon = styled.div`
  font-size: 1.5rem;
  min-width: 24px;
`;

const FileDetails = styled.div`
  flex: 1;
`;

const FileName = styled.div`
  font-weight: 500;
  color: #212529;
  margin-bottom: 0.25rem;
`;

const FileStatus = styled.div<{ completed?: boolean }>`
  font-size: 0.875rem;
  color: ${({ completed }) => completed ? '#28a745' : '#6c757d'};
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

const LoadingSpinner = styled.div`
  width: 12px;
  height: 12px;
  border: 2px solid #e9ecef;
  border-top: 2px solid #007bff;
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
`;

const InfoSection = styled.div`
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 1.5rem;
`;

const InfoText = styled.p`
  color: #6c757d;
  font-size: 0.875rem;
  line-height: 1.5;
  margin: 0 0 0.75rem 0;
  
  &:last-child {
    margin-bottom: 0;
  }
`;