import { StopOutlined } from '@ant-design/icons';
import { useIntl } from '@umijs/max';
import { Avatar, Tooltip } from 'antd';
import React from 'react';
import { getColorPrimary } from '@/utils/global';

interface UserProps {
  username: string;
  enable: boolean | undefined;
  nickname?: string;
  avatar?: string;
  jobNumber?: string;
}

export const SystemUser: React.FC<UserProps> = (props) => {
  const intl = useIntl();
  const colorPrimary = getColorPrimary();
  return (
    <span style={{ color: colorPrimary }}>
      {props.avatar && (
        <>
          <Avatar src={props.avatar} size={24} />
          &nbsp;&nbsp;
        </>
      )}
      {props.username}
      {props.nickname ? '/' + props.nickname : ''}
      {props.jobNumber ? '/' + props.jobNumber : ''}
      {!props.enable && (
        <Tooltip
          color={colorPrimary}
          placement="bottom"
          title={intl.formatMessage({ id: 'pages.operation.disable' })}
        >
          &nbsp;&nbsp;<StopOutlined style={{ color: 'red' }} />
        </Tooltip>
      )}
    </span>
  );
};

export default SystemUser;
