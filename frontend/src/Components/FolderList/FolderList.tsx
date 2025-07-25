import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { GetStoredFolders } from '../../../wailsjs/go/app/App';
import styled from 'styled-components';
import folderIcon from '../../assets/images/icons/folder-icon.svg';

interface FolderInfo {
  id: number
  name: string
  timestamp: string
  fileCount: number
}

const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch (error) {
    return timestamp
  }
};

export function FolderList() {
  const navigate = useNavigate();
  const [folders, setFolders] = useState<FolderInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFolders = async () => {
    try {
      setLoading(true);
      setError(null);
      const foldersData = await GetStoredFolders();
      setFolders(foldersData);
    } catch (err) {
      console.error('Failed to fetch folders:', err);
      setError('Failed to fetch folders. Please ensure you are logged in.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFolders();
  }, []);

  const handleFolderClick = (folderId: number) => {
    navigate(`/folder/${folderId}`);
  };

  if (loading) {
    return <FoldersContainer>Loading folders...</FoldersContainer>;
  }

  if (error) {
    return <FoldersContainer>{error}</FoldersContainer>;
  }
  
  if (!folders || folders.length === 0) {
    return (
      <FoldersContainer>
        <RefreshButton onClick={fetchFolders}>Refresh Folders</RefreshButton>
        <NoFoldersMessage>
          No folders found. Receive files via Nearby Sharing to see them organized in folders here.
        </NoFoldersMessage>
      </FoldersContainer>
    );
  }

  return (
    <FoldersContainer>
      <FoldersHeader>Received</FoldersHeader>
      
      <FoldersGrid>
        {folders.map((folder) => (
          <FolderCard 
            key={folder.id} 
            onClick={() => handleFolderClick(folder.id)}
          >
            <FolderIcon />
            <FolderInfo>
              <FolderName>{folder.name}</FolderName>
              <FolderMeta>
                <FileCount>{folder.fileCount} files</FileCount>
                <FolderDate>{formatTimestamp(folder.timestamp)}</FolderDate>
              </FolderMeta>
            </FolderInfo>
          </FolderCard>
        ))}
      </FoldersGrid>
    </FoldersContainer>
  );
}

const FoldersContainer = styled.div`
  padding: 1rem;
  margin-bottom: 1.5rem;
`;

const FoldersHeader = styled.h2`
  color: ${({ theme }) => theme.colors.darkGray};
  margin-bottom: 1rem;
`;

const FoldersGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
`;

const FolderCard = styled.div`
  display: flex;
  align-items: center;
  padding: 1rem;
  background-color: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  transition: all 0.2s ease;
  
  &:hover {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: ${({ theme }) => theme.colors.primary};
    transform: translateY(-2px);
  }
`;

const FolderIcon = styled.div`
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  background-image: url(${folderIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  margin-right: 1rem;
`;

const FolderInfo = styled.div`
  flex: 1;
  min-width: 0;
`;

const FolderName = styled.h3`
  color: ${({ theme }) => theme.colors.darkGray};
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const FolderMeta = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
`;

const FileCount = styled.span`
  color: ${({ theme }) => theme.colors.lightGray};
  font-size: 0.9rem;
  font-weight: 500;
`;

const FolderDate = styled.span`
  color: ${({ theme }) => theme.colors.lightGray};
  font-size: 0.8rem;
  opacity: 0.8;
`;

const NoFoldersMessage = styled.div`
  padding: 2rem 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.lightGray};
  background-color: rgba(255, 255, 255, 0.02);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  border: 1px dashed rgba(255, 255, 255, 0.2);
`;

const RefreshButton = styled.button`
  padding: 0.5rem 1rem;
  background-color: ${({ theme }) => theme.colors.primary};
  color: white;
  border: none;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  
  &:hover {
    background-color: #1e56c9;
  }
`;

export default FolderList;