import { CopyOutlined, ExpandOutlined, LinkOutlined } from '@ant-design/icons';
import { useIntl } from '@umijs/max';
import { Button, Card, Col, Empty, Modal, Row, Space, Tag, Typography, message } from 'antd';
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

const buildPreviewURL = (raw?: string) => {
  const endpoint = normalizeURL(raw);
  if (!endpoint) return '';
  return endpoint;
};

const maskToken = (raw?: string) => {
  const value = String(raw || '').trim();
  if (!value) return '';
  if (value.length <= 8) {
    return `${value.slice(0, 2)}***${value.slice(-2)}`;
  }
  return `${value.slice(0, 4)}***${value.slice(-4)}`;
};

const buildKey = (item: InstanceTelemetry) => `${item.namespace || 'ns'}-${item.name || 'name'}`;

const OpenClawPreviewGrid: React.FC<Props> = ({ instances }) => {
  const intl = useIntl();
  const [active, setActive] = useState<InstanceTelemetry | null>(null);

  const items = useMemo(
    () => instances.filter((item) => String(item.accessible) === 'true' || item.accessible),
    [instances],
  );
  const activeEndpoint = active ? buildPreviewURL(active.endpoint) : '';
  const activePreviewable = active ? isPreviewable(activeEndpoint) : false;

  return (
    <>
      <Card size="small" title={intl.formatMessage({ id: 'pages.openclaw.preview.title' })} className={styles.card}>
        {items.length === 0 ? (
          <Empty description={intl.formatMessage({ id: 'pages.openclaw.preview.empty' })} />
        ) : (
          <Row gutter={[12, 12]}>
            {items.map((item) => {
              const endpoint = buildPreviewURL(item.endpoint);
              const previewable = isPreviewable(endpoint);
              const gatewayToken = String(item.gatewayToken || '').trim();
              return (
                <Col xs={24} md={12} xl={8} key={buildKey(item)}>
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
                        <iframe
                          className={styles.frame}
                          src={endpoint}
                          title={`preview-${buildKey(item)}`}
                          loading="lazy"
                          sandbox="allow-forms allow-modals allow-pointer-lock allow-popups allow-same-origin allow-scripts"
                        />
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={intl.formatMessage({ id: 'pages.openclaw.preview.endpoint.missing' })} />
                      )}
                    </div>
                    {gatewayToken ? (
                      <div className={styles.tokenRow}>
                        <Typography.Text type="secondary" className={styles.tokenText}>
                          {intl.formatMessage({ id: 'pages.openclaw.preview.token.label' })}: {maskToken(gatewayToken)}
                        </Typography.Text>
                        <Button
                          size="small"
                          icon={<CopyOutlined />}
                          className={styles.tokenCopyButton}
                          onClick={async () => {
                            try {
                              await navigator.clipboard.writeText(gatewayToken);
                              message.success(intl.formatMessage({ id: 'pages.copy.success' }));
                            } catch (_err) {
                              message.error(intl.formatMessage({ id: 'pages.copy.failed' }));
                            }
                          }}
                        >
                          {intl.formatMessage({ id: 'pages.openclaw.preview.token.copy' })}
                        </Button>
                      </div>
                    ) : null}
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
