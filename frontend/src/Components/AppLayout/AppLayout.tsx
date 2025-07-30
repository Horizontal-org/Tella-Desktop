import React from 'react';
import styled from 'styled-components';
import { Sidebar } from '../Sidebar/Sidebar';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ 
  children,
}: AppLayoutProps) {
  return (
    <LayoutContainer>
      <Sidebar />
      
      <MainContent>
        <ContentSection>
          {children}
        </ContentSection>
      </MainContent>
    </LayoutContainer>
  );
}

const LayoutContainer = styled.div`
  display: flex;
  height: 100vh;
  background-color: ${({ theme }) => theme.colors.background};
`;

const MainContent = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
`;

const ContentSection = styled.div`
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
`;