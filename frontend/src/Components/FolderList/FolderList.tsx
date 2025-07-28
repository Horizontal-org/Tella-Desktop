import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { GetStoredFolders } from '../../../wailsjs/go/app/App';
import {
  FoldersContainer,
  FoldersHeader,
  HeaderTitle,
  ToolbarContainer,
  ToolbarActions,
  ExportButton,
  DeleteButton,
  ExportIcon,
  DeleteIcon,
  TableContainer,
  Table,
  TableHeader,
  TableBody,
  HeaderRow,
  TableRow,
  CheckboxCell,
  NameCell,
  FilesCell,
  DateCell,
  NameHeader,
  FilesHeader,
  DateHeader,
  FolderIcon,
  FolderName,
  Checkbox,
  NoFoldersMessage,
  RefreshButton
} from './FolderListStyles';

interface FolderInfo {
  id: number
  name: string
  timestamp: string
  fileCount: number
}

const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp);
    return date.toLocaleDateString('en-US', {
      day: 'numeric',
      month: 'short',
      year: 'numeric'
    });
  } catch (error) {
    return timestamp
  }
};

export function FolderList() {
  const navigate = useNavigate();
  const [folders, setFolders] = useState<FolderInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedFolders, setSelectedFolders] = useState<Set<number>>(new Set());
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

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setSelectedFolders(new Set());
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleFolderClick = (folderId: number, event: React.MouseEvent) => {
    if (selectedFolders.has(folderId) && selectedFolders.size === 1) {
      setSelectedFolders(new Set());
    } else {
      setSelectedFolders(new Set([folderId]));
    }
  };

  const handleFolderDoubleClick = (folderId: number) => {
    navigate(`/folder/${folderId}`);
  };

  const handleCheckboxChange = (folderId: number, checked: boolean) => {
    const newSelected = new Set(selectedFolders);
    if (checked) {
      newSelected.add(folderId);
    } else {
      newSelected.delete(folderId);
    }
    setSelectedFolders(newSelected);
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedFolders(new Set(folders.map(f => f.id)));
    } else {
      setSelectedFolders(new Set());
    }
  };

  const isAllSelected = folders.length > 0 && selectedFolders.size === folders.length;
  const isIndeterminate = selectedFolders.size > 0 && selectedFolders.size < folders.length;

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
      <FoldersHeader>
        <HeaderTitle>Received</HeaderTitle>
      </FoldersHeader>
      
      <ToolbarContainer $isVisible={selectedFolders.size > 0}>
        <ToolbarActions>
          <ExportButton>
            <ExportIcon />
            EXPORT AS ZIP
          </ExportButton>
          <DeleteButton>
            <DeleteIcon />
            DELETE
          </DeleteButton>
        </ToolbarActions>
      </ToolbarContainer>
      
      <TableContainer>
        <Table>
          <TableHeader>
            <HeaderRow>
              <CheckboxCell>
                <Checkbox
                  type="checkbox"
                  checked={isAllSelected}
                  ref={(input) => {
                    if (input) input.indeterminate = isIndeterminate;
                  }}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              </CheckboxCell>
              <NameHeader>Name</NameHeader>
              <FilesHeader>Files</FilesHeader>
              <DateHeader>Date received</DateHeader>
            </HeaderRow>
          </TableHeader>
          <TableBody>
            {folders.map((folder) => (
              <TableRow
                key={folder.id}
                $isSelected={selectedFolders.has(folder.id)}
                onClick={(e) => handleFolderClick(folder.id, e)}
                onDoubleClick={() => handleFolderDoubleClick(folder.id)}
              >
                <CheckboxCell>
                  <Checkbox
                    type="checkbox"
                    checked={selectedFolders.has(folder.id)}
                    onChange={(e) => {
                      e.stopPropagation();
                      handleCheckboxChange(folder.id, e.target.checked);
                    }}
                  />
                </CheckboxCell>
                <NameCell>
                  <FolderIcon />
                  <FolderName>{folder.name}</FolderName>
                </NameCell>
                <FilesCell>{folder.fileCount} files</FilesCell>
                <DateCell>{formatTimestamp(folder.timestamp)}</DateCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </FoldersContainer>
  );
}

export default FolderList;