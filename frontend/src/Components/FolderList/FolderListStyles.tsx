import styled from 'styled-components';
import folderIcon from '../../assets/images/icons/folder-icon.svg';
import trashIcon from '../../assets/images/icons/trash-icon.svg'
import zipIcon from '../../assets/images/icons/zip-icon.svg'

export const FoldersContainer = styled.div`
  padding: 1rem;
  margin-bottom: 1.5rem;
`;

export const FoldersHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
`;

export const HeaderTitle = styled.h2`
  color: ${({ theme }) => theme.colors.darkGray};
  margin: 0;
`;

export const ToolbarContainer = styled.div<{ $isVisible: boolean }>`
  height: 60px;
  padding: 0.75rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  visibility: ${({ $isVisible }) => $isVisible ? 'visible' : 'hidden'};
  opacity: ${({ $isVisible }) => $isVisible ? 1 : 0};
  transition: opacity 0.2s ease;
`;

export const ToolbarActions = styled.div`
  display: flex;
  gap: 0.75rem;
  align-items: center;
`;

export const ExportButton = styled.button`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem;
  background-color: transparent;
  color: #8B8E8F;
  border: 2px solid #D9D9D9;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
`;

export const DeleteButton = styled.button`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem;
  background-color: transparent;
  color: #8B8E8F;
  border: 2px solid #D9D9D9;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
`;

export const ExportIcon = styled.div`
  width: 16px;
  height: 16px;
  background-color: currentColor;
  mask: url(${zipIcon}) no-repeat center;
  mask-size: contain;
`;

export const DeleteIcon = styled.div`
  width: 16px;
  height: 16px;
  background-color: currentColor;
  mask: url(${trashIcon}) no-repeat center;
  mask-size: contain;
`;

export const TableContainer = styled.div`
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  overflow: hidden;
`;

export const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
`;

export const TableHeader = styled.thead`
  background-color: rgba(255, 255, 255, 0.05);
`;

export const TableBody = styled.tbody``;

export const HeaderRow = styled.tr``;

export const TableRow = styled.tr<{ $isSelected?: boolean }>`
  cursor: pointer;
  transition: all 0.2s ease;
  
  background-color: ${({ $isSelected }) => 
    $isSelected ? 'rgba(59, 130, 246, 0.15)' : 'transparent'
  };
  
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  
  &:hover {
    background-color: ${({ $isSelected }) => 
      $isSelected ? 'rgba(59, 130, 246, 0.2)' : 'rgba(255, 255, 255, 0.05)'
    };
  }
`;

export const BaseCell = styled.td`
  padding: 1rem 0.75rem;
  color: ${({ theme }) => theme.colors.lightGray};
  vertical-align: middle;
`;

export const BaseHeader = styled.th`
  padding: 1rem 0.75rem;
  text-align: left;
  color: ${({ theme }) => theme.colors.darkGray};
  font-weight: 600;
  font-size: 0.875rem;
  border-bottom: 2px solid rgba(255, 255, 255, 0.1);
`;

export const CheckboxCell = styled(BaseCell)`
  width: 50px;
  padding-left: 1rem;
`;

export const NameCell = styled(BaseCell)`
  display: flex;
  align-items: center;
  gap: 0.75rem;
`;

export const FilesCell = styled(BaseCell)`
  font-size: 0.875rem;
`;

export const DateCell = styled(BaseCell)`
  font-size: 0.875rem;
`;

export const NameHeader = styled(BaseHeader)``;
export const FilesHeader = styled(BaseHeader)``;
export const DateHeader = styled(BaseHeader)``;

export const FolderIcon = styled.div`
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  background-image: url(${folderIcon});
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
`;

export const FolderName = styled.span`
  font-weight: 500;
  color: ${({ theme }) => theme.colors.darkGray};
`;

export const Checkbox = styled.input`
  width: 16px;
  height: 16px;
  accent-color: ${({ theme }) => theme.colors.primary};
  cursor: pointer;
`;

export const NoFoldersMessage = styled.div`
  padding: 2rem 1rem;
  text-align: center;
  color: ${({ theme }) => theme.colors.lightGray};
  background-color: rgba(255, 255, 255, 0.02);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  border: 1px dashed rgba(255, 255, 255, 0.2);
`;

export const RefreshButton = styled.button`
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