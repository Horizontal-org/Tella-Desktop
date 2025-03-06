import { useState, useEffect } from 'react';
import { GetStoredFiles } from '../../../wailsjs/go/app/App';
import styled from 'styled-components';

// Define the FileInfo type to match the struct in Go
interface FileInfo {
  name: string;
  mimeType: string;
  timestamp: string;
}

// Format the timestamp for display
const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch (error) {
    return timestamp; // Return original if parsing fails
  }
};

// Get a friendly name for the MIME type
const getFriendlyMimeType = (mimeType: string): string => {
  const mimeMap: Record<string, string> = {
    'text/plain': 'Text',
    'text/html': 'HTML',
    'text/css': 'CSS',
    'text/javascript': 'JavaScript',
    'application/json': 'JSON',
    'application/xml': 'XML',
    'application/pdf': 'PDF',
    'application/msword': 'Word',
    'application/vnd.ms-excel': 'Excel',
    'application/vnd.ms-powerpoint': 'PowerPoint',
    'image/jpeg': 'JPEG Image',
    'image/png': 'PNG Image',
    'image/gif': 'GIF Image',
    'image/svg+xml': 'SVG Image',
    'audio/mpeg': 'MP3 Audio',
    'audio/wav': 'WAV Audio',
    'video/mp4': 'MP4 Video',
    'video/mpeg': 'MPEG Video',
    'application/octet-stream': 'Binary File',
  };

  return mimeMap[mimeType] || mimeType;
};

export function FilesList() {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFiles = async () => {
    try {
      setLoading(true);
      setError(null);
      const filesData = await GetStoredFiles();
      setFiles(filesData);
    } catch (err) {
      console.error('Failed to fetch files:', err);
      setError('Failed to fetch files. Please ensure you are logged in.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFiles();
  }, []);

  if (loading) {
    return <FilesContainer>Loading files...</FilesContainer>;
  }

  if (error) {
    return <FilesContainer>{error}</FilesContainer>;
  }

  return (
    <FilesContainer>
      <FilesHeader>Stored Files</FilesHeader>
      <RefreshButton onClick={fetchFiles}>Refresh Files</RefreshButton>
      
      {files.length === 0 ? (
        <NoFilesMessage>No files found. Receive files to see them listed here.</NoFilesMessage>
      ) : (
        <FilesTable>
          <thead>
            <tr>
              <TableHeader>Name</TableHeader>
              <TableHeader>Type</TableHeader>
              <TableHeader>Date Added</TableHeader>
            </tr>
          </thead>
          <tbody>
            {files.map((file, index) => (
              <tr key={index}>
                <TableCell>{file.name}</TableCell>
                <TableCell>{getFriendlyMimeType(file.mimeType)}</TableCell>
                <TableCell>{formatTimestamp(file.timestamp)}</TableCell>
              </tr>
            ))}
          </tbody>
        </FilesTable>
      )}
    </FilesContainer>
  );
}

// Styled components
const FilesContainer = styled.div`
  padding: 1rem;
  margin-bottom: 1.5rem;
`;

const FilesHeader = styled.h2`
  color: ${({ theme }) => theme.colors.darkGray};
  margin-bottom: 1rem;
`;

const FilesTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
`;

const TableHeader = styled.th`
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid ${({ theme }) => theme.colors.lightGray};
  color: ${({ theme }) => theme.colors.darkGray};
`;

const TableCell = styled.td`
  padding: 0.75rem;
  border-bottom: 1px solid ${({ theme }) => theme.colors.lightGray};
  color: ${({ theme }) => theme.colors.lightGray};
`;

const NoFilesMessage = styled.div`
  padding: 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.lightGray};
`;

const RefreshButton = styled.button`
  padding: 0.5rem 1rem;
  background-color: ${({ theme }) => theme.colors.primary};
  color: white;
  border: none;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  margin-bottom: 1rem;
  
  &:hover {
    background-color: #1e56c9;
  }
`;

export default FilesList;