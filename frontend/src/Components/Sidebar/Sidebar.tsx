import { useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';

interface SidebarProps {
  className?: string;
}

export function Sidebar({ className }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();

  const handleNavigation = (path: string) => {
    navigate(path);
  };

  const isActive = (path: string) => {
    if (path === '/' && (location.pathname === '/' || location.pathname.startsWith('/folder'))) return true;
    if (path !== '/' && location.pathname.startsWith(path)) return true;
    return false;
  };

  return (
    <SidebarContainer className={className}>
      <SidebarHeader>
        <AppTitle>Tella Desktop</AppTitle>
      </SidebarHeader>
      
      <Navigation>
        <NavItem 
          $isActive={isActive('/')} 
          onClick={() => handleNavigation('/')}
        >
          <NavText>Received</NavText>
        </NavItem>
        
        <NavItem 
          $isActive={isActive('/nearby-sharing')} 
          onClick={() => handleNavigation('/nearby-sharing')}
        >
          <NavText>Nearby Sharing</NavText>
        </NavItem>
      </Navigation>
    </SidebarContainer>
  );
}

const SidebarContainer = styled.div`
  width: 280px;
  background-color: #f8f9fa;
  border-right: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
`;

const SidebarHeader = styled.div`
  padding: 2rem 1.5rem 1.5rem;
  border-bottom: 1px solid #e9ecef;
`;

const AppTitle = styled.h1`
  color: #212529;
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
`;

const Navigation = styled.nav`
  padding: 1rem 0;
  flex: 1;
`;

const NavItem = styled.div<{ $isActive: boolean }>`
  display: flex;
  align-items: center;
  padding: 1rem 1.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 0;
  margin: 0 0.75rem;
  border-radius: ${({ theme }) => theme.borderRadius.default};
  
  background-color: ${({ $isActive }) => $isActive ? '#065485' : 'transparent'};
  color: ${({ $isActive }) => $isActive ? 'white' : '#6c757d'};
  
  &:hover {
    background-color: ${({ $isActive }) => $isActive ? '#054a7a' : '#e9ecef'};
    transform: ${({ $isActive }) => $isActive ? 'none' : 'translateX(4px)'};
  }
`;

const NavText = styled.span`
  font-size: 1rem;
  font-weight: 500;
`;