import { useNavigate } from 'react-router-dom';
import { FolderList } from '../FolderList/FolderList';
import styled from 'styled-components';

export function Dashboard() {
  const navigate = useNavigate();

  const handleNearbySharing = () => {
    navigate('/nearby-sharing');
  };

  return (
    <DashboardContainer>
      <DashboardHeader>
        <NearbyButton onClick={handleNearbySharing}>
          Nearby Sharing
        </NearbyButton>
      </DashboardHeader>
      
      <ContentSection>
        <FolderList />
      </ContentSection>
    </DashboardContainer>
  );
}

const DashboardContainer = styled.div`
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
`;

const DashboardHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid ${({ theme }) => theme.colors.lightGray};
`;

const Title = styled.h1`
  color: ${({ theme }) => theme.colors.darkGray};
  margin: 0;
  font-size: ${({ theme }) => theme.fontSizes.xxlarge};
`;

const NearbyButton = styled.button`
  padding: 0.75rem 1.5rem;
  background-color: ${({ theme }) => theme.colors.primary};
  color: white;
  border: none;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  font-size: ${({ theme }) => theme.fontSizes.medium};
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: #1e56c9;
  }
`;

const ContentSection = styled.div`
  background-color: rgba(255, 255, 255, 0.05);
  border-radius: ${({ theme }) => theme.borderRadius.default};
  padding: 1.5rem;
`;