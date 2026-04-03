import { QuestionCircleOutlined } from '@ant-design/icons';
import { FormattedMessage } from '@umijs/max';
import { Tooltip } from 'antd';
import React from 'react';
import { getColorPrimary } from '@/utils/global';
export const About: React.FC = () => {
  const colorPrimary = getColorPrimary();
  return (
    <Tooltip
      color={colorPrimary}
      title={<FormattedMessage id="app.description" />}
    >
      <QuestionCircleOutlined />
    </Tooltip>
  );
};
