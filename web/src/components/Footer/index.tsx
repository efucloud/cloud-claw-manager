import { GithubOutlined } from '@ant-design/icons';
import { DefaultFooter } from '@ant-design/pro-components';
import { useIntl } from '@umijs/max';
import React from 'react';

const Footer: React.FC = () => {
  const intl = useIntl()
  return (
    <DefaultFooter
      style={{
        background: 'none',
      }}
      copyright="2025"
      links={[
        {
          key: 'efucloud',
          title: intl.formatMessage({ id: 'company.efucloud' }),
          href: 'https://efucloud.com',
          blankTarget: true,
        },
        {
          key: 'github',
          title: <GithubOutlined />,
          href: 'https://github.com/efucloud',
          blankTarget: true,
        },
      ]}
    />
  );
};

export default Footer;
