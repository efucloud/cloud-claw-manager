import { PageContainer } from '@ant-design/pro-components';
import { createStyles } from 'antd-style';
import React, { useEffect } from 'react';
import { Footer } from '@/components';
import type { AccessTokenResponse, LoginByOIDC } from '@/services/common.d';
import { loginByOidc } from '@/services/oauth.api';
import { addToken, getSearchParams } from '@/utils/global';

const useStyles = createStyles(({ token }) => {
  return {
    container: {
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      overflow: 'auto',
      backgroundColor: token.colorBgContainer,
    },
  };
});

const OAuthCallback: React.FC = () => {
  const { styles } = useStyles();
  const code = getSearchParams().get('code') as string;
  const login = async () => {
    const loginParam = {} as LoginByOIDC;
    loginParam.code = code;
    loginParam.redirectUri = window.location.origin + window.location.pathname;
    const token = (await loginByOidc(loginParam)) as AccessTokenResponse;
    addToken(token);
    window.location.href = '/dashboard';
  };

  useEffect(() => {
    login();
  }, []);
  return (
    <PageContainer>
      <div className={styles.container}></div>
      <Footer />
    </PageContainer>
  );
};

export default OAuthCallback;
