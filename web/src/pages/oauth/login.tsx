import { PageContainer } from '@ant-design/pro-components';
import { nanoid } from 'nanoid';
import React, { useEffect, useState } from 'react';
import { Footer } from '@/components';
import type { OidcConfig } from '@/services/common.d';
import { getAuthorizeInfo } from '@/services/oauth.api';
import { getToken } from '@/utils/global';

const LoginIndex: React.FC = () => {
  const [oauthConfig, setOauthConfig] = useState<OidcConfig>();
  const getCode = async () => {
    const res = (await getAuthorizeInfo()) as OidcConfig;
    setOauthConfig(res);
  };
  useEffect(() => {
    getCode();
  }, []);
  useEffect(() => {
    const token = getToken();
    if (token.access_token) {
      window.location.href = '/dashboard';
    } else if (oauthConfig) {
      const redirect_uri = `${window.location.origin}/oauth/callback`;
      const state = nanoid(20);
      window.location.href = `${oauthConfig.issuer}/oauth/authorize?redirect_uri=${redirect_uri}&client_id=${oauthConfig.clientId}&response_type=code&state=${state}`;
    }
  }, [oauthConfig]);
  return (
    <PageContainer>
      <div></div>
      <Footer />
    </PageContainer>
  );
};

export default LoginIndex;
