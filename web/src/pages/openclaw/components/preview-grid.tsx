import { ExpandOutlined, LinkOutlined } from '@ant-design/icons';
import { useIntl } from '@umijs/max';
import { Button, Card, Col, Empty, Modal, Row, Space, Tag } from 'antd';
import React, { useMemo, useState } from 'react';
import type { InstanceTelemetry } from '../types';
import styles from './preview-grid.less';

type Props = {
  instances: InstanceTelemetry[];
};

const normalizeURL = (raw?: string) => {
  const value = String(raw || '').trim();
  if (!value) return '';
  if (/^https?:\/\//i.test(value)) return value;
  return `https://${value}`;
};

const isPreviewable = (raw?: string) => {
  const value = normalizeURL(raw);
  if (!value) return false;
  try {
    const u = new URL(value);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch (_err) {
    return false;
  }
};

const buildPreviewURL = (raw?: string, token?: string) => {
  const endpoint = normalizeURL(raw);
  if (!endpoint) return '';
  const tokenValue = String(token || '').trim();
  if (!tokenValue) {
    return endpoint;
  }
  try {
    const url = new URL(endpoint);
    url.searchParams.set('token', tokenValue);
    return url.toString();
  } catch (_err) {
    const sep = endpoint.includes('?') ? '&' : '?';
    return `${endpoint}${sep}token=${encodeURIComponent(tokenValue)}`;
  }
};

const buildKey = (item: InstanceTelemetry) => `${item.namespace || 'ns'}-${item.name || 'name'}`;

const OpenClawPreviewGrid: React.FC<Props> = ({ instances }) => {
  const intl = useIntl();
  const [active, setActive] = useState<InstanceTelemetry | null>(null);

  const items = useMemo(
    () => instances.filter((item) => String(item.accessible) === 'true' || item.accessible),
    [instances],
  );
  const activeEndpoint = active ? buildPreviewURL(active.endpoint, active.gatewayToken) : '';
  const activePreviewable = active ? isPreviewable(activeEndpoint) : false;

  return (
    <>
      <Card size="small" title={intl.formatMessage({ id: 'pages.openclaw.preview.title' })} className={styles.card}>
        {items.length === 0 ? (
          <Empty description={intl.formatMessage({ id: 'pages.openclaw.preview.empty' })} />
        ) : (
          <Row gutter={[12, 12]}>
            {items.map((item) => {
              const endpoint = buildPreviewURL(item.endpoint, item.gatewayToken);
              const previewable = isPreviewable(endpoint);
              return (
                <Col xs={24} md={12} lg={12} xl={12} key={buildKey(item)}>
                  <Card
                    size="small"
                    className={styles.instanceCard}
                    title={item.displayName || item.name || '-'}
                    extra={
                      <Space size={8} wrap className={styles.cardExtra}>
                        <Tag color={item.readyPods && item.readyPods > 0 ? 'success' : 'default'}>
                          {item.readyPods && item.readyPods > 0
                            ? intl.formatMessage({ id: 'pages.openclaw.preview.running' })
                            : intl.formatMessage({ id: 'pages.openclaw.preview.stopped' })}
                        </Tag>
                        <Button
                          size="small"
                          icon={<ExpandOutlined />}
                          disabled={!previewable}
                          onClick={() => setActive(item)}
                        >
                          {intl.formatMessage({ id: 'pages.openclaw.preview.fullscreen' })}
                        </Button>
                        <Button
                          size="small"
                          icon={<LinkOutlined />}
                          disabled={!previewable}
                          href={endpoint || undefined}
                          target="_blank"
                          rel="noreferrer"
                        >
                          {intl.formatMessage({ id: 'pages.openclaw.preview.newWindow' })}
                        </Button>
                      </Space>
                    }
                  >
                    <div className={styles.frameWrap}>
                      {previewable ? (
                        <div className={styles.frameViewport}>
                          <iframe
                            className={styles.frame}
                            src={endpoint}
                            title={`preview-${buildKey(item)}`}
                            loading="lazy"
                            sandbox="allow-forms allow-modals allow-pointer-lock allow-popups allow-same-origin allow-scripts"
                          />
                        </div>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={intl.formatMessage({ id: 'pages.openclaw.preview.endpoint.missing' })} />
                      )}
                    </div>
                  </Card>
                </Col>
              );
            })}
          </Row>
        )}
      </Card>

      <Modal
        open={Boolean(active)}
        className={styles.fullscreenModal}
        title={
          active
            ? intl.formatMessage(
                { id: 'pages.openclaw.preview.fullscreen.title' },
                { name: active.displayName || active.name || '-' },
              )
            : intl.formatMessage({ id: 'pages.openclaw.preview.fullscreen' })
        }
        width="100vw"
        footer={
          activePreviewable ? (
            <Button icon={<LinkOutlined />} href={activeEndpoint} target="_blank" rel="noreferrer">
              {intl.formatMessage({ id: 'pages.openclaw.preview.newWindow' })}
            </Button>
          ) : null
        }
        destroyOnClose
        style={{ top: 0, paddingBottom: 0, maxWidth: '100vw' }}
        bodyStyle={{ height: 'calc(100vh - 55px)' }}
        onCancel={() => setActive(null)}
      >
        {active && activePreviewable ? (
          <iframe
            className={styles.fullscreenFrame}
            src={activeEndpoint}
            title={`fullscreen-${buildKey(active)}`}
            sandbox="allow-forms allow-modals allow-pointer-lock allow-popups allow-same-origin allow-scripts"
          />
        ) : (
          <Empty description={intl.formatMessage({ id: 'pages.openclaw.preview.address.unavailable' })} />
        )}
      </Modal>
    </>
  );
};

export default OpenClawPreviewGrid;
