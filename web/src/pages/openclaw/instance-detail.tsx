import { openclawDashboard } from '@/services/openclaw.api';
import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history, useIntl, useParams } from '@umijs/max';
import { Button, Card, Col, Descriptions, Empty, Row, Spin, Statistic, Tag, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  decodePart,
  formatBytes,
  formatNumber,
  normalizeProvider,
} from './format';
import type { DashboardResponse, InstanceTelemetry } from './types';
import styles from './dashboard.less';

const statusColor = (value?: string) => {
  const text = String(value || '').trim().toLowerCase();
  if (text === 'ok' || text === 'running' || text === 'active') {
    return 'success';
  }
  if (text === 'warn' || text === 'warning') {
    return 'warning';
  }
  if (text === 'critical' || text === 'error' || text === 'failed' || text === 'stopped') {
    return 'error';
  }
  return 'default';
};

const OpenClawInstanceDetailPage: React.FC = () => {
  const intl = useIntl();
  const params = useParams<{ namespace?: string; name?: string }>();
  const namespace = decodePart(params?.namespace);
  const name = decodePart(params?.name);
  const [loading, setLoading] = useState<boolean>(false);
  const [instance, setInstance] = useState<InstanceTelemetry | null>(null);

  const ready = Number(instance?.readyPods || 0) > 0;
  const runtime = instance?.runtimeInsights;

  const loadData = useCallback(async () => {
    if (!namespace || !name) {
      setInstance(null);
      return;
    }
    setLoading(true);
    try {
      const resp = await openclawDashboard<DashboardResponse>({
        namespace,
        includeGatewayToken: false,
      });
      const payload = (resp as any)?.data || resp || {};
      const rows: InstanceTelemetry[] = payload.instances || [];
      const matched = rows.find((item) => item.namespace === namespace && item.name === name) || null;
      setInstance(matched);
    } catch (_error) {
      message.error(intl.formatMessage({ id: 'pages.openclaw.dashboard.load.failed' }));
      setInstance(null);
    } finally {
      setLoading(false);
    }
  }, [intl, name, namespace]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const pageTitle = useMemo(() => {
    return instance?.displayName || instance?.name || name || intl.formatMessage({ id: 'pages.openclaw.instance.title' });
  }, [instance?.displayName, instance?.name, intl, name]);

  return (
    <PageContainer
      title={pageTitle}
      subTitle={namespace && name ? `${namespace}/${name}` : undefined}
      extra={[
        <Button key="refresh" icon={<ReloadOutlined />} onClick={() => void loadData()}>
          {intl.formatMessage({ id: 'pages.operation.refresh' })}
        </Button>,
        <Button key="back" icon={<ArrowLeftOutlined />} onClick={() => history.push('/dashboard')}>
          {intl.formatMessage({ id: 'pages.operation.back' })}
        </Button>,
      ]}
    >
      <div className={styles.wrapper}>
        <Spin spinning={loading}>
          {!namespace || !name ? (
            <div className={styles.emptyWrap}>
              <Empty description={intl.formatMessage({ id: 'pages.openclaw.instance.invalid' })} />
            </div>
          ) : !instance ? (
            <div className={styles.emptyWrap}>
              <Empty description={intl.formatMessage({ id: 'pages.openclaw.instance.notFound' })} />
            </div>
          ) : (
            <>
              <Row gutter={[16, 16]} className={styles.statRow}>
                <Col xs={12} sm={12} md={6} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.dashboard.state.ready' })}
                      value={ready ? intl.formatMessage({ id: 'pages.openclaw.value.yes' }) : intl.formatMessage({ id: 'pages.openclaw.value.no' })}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={12} md={6} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.tokens.input' })}
                      value={formatNumber(instance.inputTokens24h, 0)}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={12} md={6} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.tokens.output' })}
                      value={formatNumber(instance.outputTokens24h, 0)}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={12} md={6} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.cost' })}
                      value={formatNumber(instance.costUSD24h, 4)}
                    />
                  </Card>
                </Col>
              </Row>

              <Card className={styles.detailCard} title={intl.formatMessage({ id: 'pages.openclaw.instance.basic' })}>
                <Descriptions column={{ xs: 1, sm: 1, md: 2, lg: 3 }}>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.dashboard.field.namespace' })}>
                    {instance.namespace || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.displayName' })}>
                    {instance.displayName || instance.name || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.ownerUser' })}>
                    {instance.ownerUsername || instance.ownerEmail || instance.ownerId || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.purpose' })}>
                    {instance.purpose || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.provider' })}>
                    {normalizeProvider(instance.provider) || intl.formatMessage({ id: 'pages.openclaw.value.unknown' })}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.memory' })}>
                    {formatBytes(instance.memoryBytes)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.tokens' })}>
                    {`${formatNumber(instance.inputTokens24h, 0)} / ${formatNumber(instance.outputTokens24h, 0)}`}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.field.cost' })}>
                    {formatNumber(instance.costUSD24h, 4)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.dashboard.field.updatedAt' })}>
                    {instance.updatedAt || '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>

              <Card className={styles.detailCard} title={intl.formatMessage({ id: 'pages.openclaw.instance.runtime' })}>
                <Descriptions column={{ xs: 1, sm: 1, md: 2, lg: 3 }}>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.status' })}>
                    <Tag color={statusColor(runtime?.statusState)}>{runtime?.statusState || '-'}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.gateway' })}>
                    <Tag color={statusColor(runtime?.gatewayState)}>{runtime?.gatewayState || '-'}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.security' })}>
                    <Tag color={statusColor(runtime?.securityState)}>{runtime?.securityState || '-'}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.version' })}>
                    {runtime?.runtimeVersion || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.workspace' })}>
                    {runtime?.openclawWorkspace || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.model' })}>
                    {runtime?.openclawPrimaryModel || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.agents' })}>
                    {formatNumber(runtime?.openclawAgentsCount, 0)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.providers' })}>
                    {formatNumber(runtime?.openclawProvidersCount, 0)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.cron' })}>
                    {formatNumber(runtime?.cronJobsCount, 0)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.sessionIndexes' })}>
                    {formatNumber(runtime?.sessionIndexFilesCount, 0)}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.disableAuth' })}>
                    {runtime?.controlUiDisableDeviceAuth === undefined
                      ? '-'
                      : runtime.controlUiDisableDeviceAuth
                        ? intl.formatMessage({ id: 'pages.openclaw.value.yes' })
                        : intl.formatMessage({ id: 'pages.openclaw.value.no' })}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.allowedOrigins' })}>
                    {(runtime?.controlUiAllowedOrigins || []).length > 0
                      ? (runtime?.controlUiAllowedOrigins || []).join(', ')
                      : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.source' })}>
                    {runtime?.sourcePod && runtime?.sourceContainer
                      ? `${runtime.sourcePod}/${runtime.sourceContainer}`
                      : runtime?.sourcePod || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.collectedAt' })}>
                    {runtime?.collectedAt || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item
                    label={intl.formatMessage({ id: 'pages.openclaw.instance.runtime.lastError' })}
                    span={2}
                  >
                    {runtime?.lastError || '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </>
          )}
        </Spin>
      </div>
    </PageContainer>
  );
};

export default OpenClawInstanceDetailPage;
