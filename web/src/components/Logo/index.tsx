import {
  AndroidOutlined,
  AppleOutlined,
  ChromeOutlined,
  GithubOutlined,
  GitlabOutlined,
  HarmonyOSOutlined,
  LinuxOutlined,
  WechatOutlined,
  WindowsOutlined,
} from '@ant-design/icons';
import { Image } from 'antd';
import React from 'react';
import { Nvidia } from '../icons';

interface LogoProps {
  name: string | undefined; // logo名称
  size?: number; // 高度
}

export const Logo: React.FC<LogoProps> = (props) => {
  if (props.size === 0) {
    props.size = 25;
  }
  if (props.name === undefined) {
    return <p style={{ fontSize: props.size }}>{props.name}</p>;
  }
  let node = {} as React.ReactNode;
  if (props.name.indexOf('Chrome') > -1) {
    node = <ChromeOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('HarmonyOS') > -1) {
    node = <HarmonyOSOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Android') > -1) {
    node = <AndroidOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Wechat') > -1) {
    node = (
      <WechatOutlined
        style={{ fontSize: props.size, color: 'rgb(51, 204, 0)' }}
      />
    );
  } else if (props.name.indexOf('Windows') > -1) {
    node = <WindowsOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Linux') > -1) {
    node = <LinuxOutlined style={{ fontSize: props.size }} />;
  } else if (
    props.name.indexOf('Apple') > -1 ||
    props.name.indexOf('MacIntel') > -1 ||
    props.name.indexOf('MacArm') > -1
  ) {
    node = <AppleOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Github') > -1) {
    node = <GithubOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Gitlab') > -1) {
    node = <GitlabOutlined style={{ fontSize: props.size }} />;
  } else if (props.name.indexOf('Safari') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/safari.png"
        alt="safari"
      />
    );
  } else if (props.name.indexOf('UCBrowser') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/uc.jpg"
        alt="safari"
      />
    );
  } else if (props.name.indexOf('Opera') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/opera.png"
        alt="opera"
      />
    );
  } else if (props.name.indexOf('Edge') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/edge.webp"
        alt="edge"
      />
    );
  } else if (props.name.indexOf('Firefox') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/firefox.webp"
        alt="firefox"
      />
    );
  } else if (props.name.indexOf('nvidia') > -1) {
    node = <Nvidia />;
  } else if (props.name.indexOf('hygon') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/hygon.png"
        alt="hygon"
      />
    );
  } else if (props.name.indexOf('ascend') > -1) {
    node = (
      <Image
        preview={false}
        width={props.size}
        src="/logo/ascend.svg"
        alt="ascend"
      />
    );
  } else {
    node = <p>{props.name}</p>;
  }
  return node;
};

export default Logo;
