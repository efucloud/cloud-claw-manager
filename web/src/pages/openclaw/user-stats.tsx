import { openclawDashboard } from '@/services/openclaw.api';
import { ArrowLeftOutlined, EyeOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history, useIntl } from '@umijs/max';
import { Button, Card, Col, Empty, Row, Space, Spin, Statistic, Tag, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { toNumber } from './format';
import type { DashboardResponse } from './types';
import styles from './dashboard.less';

const isOkState = (value?: string) => String(value || '').trim().toLowerCase() === 'ok';

const stateTagColor = (value?: string) => {
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

const OpenClawUserStatsPage: React.FC = () => {
  const intl = useIntl();
  const [loading, setLoading] = useState<boolean>(false);
  const [data, setData] = useState<DashboardResponse>({});

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await openclawDashboard<DashboardResponse>({ includeGatewayToken: false });
      const payload = (resp as any)?.data || resp;
      setData(payload || {});
    } catch (_error) {
      message.error(intl.formatMessage({ id: 'pages.openclaw.dashboard.load.failed' }));
    } finally {
      setLoading(false);
    }
  }, [intl]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const instanceRows = data.instances || [];
  const runningInstances = useMemo(
    () =>
      instanceRows.filter(
        (row) => Number(row.readyPods || 0) > 0 && Boolean(String(row.endpoint || '').trim()),
      ).length,
    [instanceRows],
  );
  const runtimeSummary = useMemo(() => {
    let statusOk = 0;
    let gatewayOk = 0;
    let securityOk = 0;
    instanceRows.forEach((item) => {
      if (String(item.runtimeInsights?.statusState || '').toLowerCase() === 'ok') {
        statusOk += 1;
      }
      if (String(item.runtimeInsights?.gatewayState || '').toLowerCase() === 'ok') {
        gatewayOk += 1;
      }
      if (String(item.runtimeInsights?.securityState || '').toLowerCase() === 'ok') {
        securityOk += 1;
      }
    });
    return {
      statusOk,
      gatewayOk,
      securityOk,
    };
  }, [instanceRows]);
  const usageSummary = useMemo(() => {
    return instanceRows.reduce(
      (acc, row) => {
        acc.input += toNumber(row.inputTokens24h);
        acc.output += toNumber(row.outputTokens24h);
        acc.cost += toNumber(row.costUSD24h);
        return acc;
      },
      { input: 0, output: 0, cost: 0 },
    );
  }, [instanceRows]);
  const interventionRows = useMemo(() => {
    const rows = instanceRows
      .map((item) => {
        const reasons: string[] = [];
        if (toNumber(item.readyPods) <= 0) {
          reasons.push(intl.formatMessage({ id: 'pages.openclaw.userStats.reason.notReady' }));
        }
        if (!String(item.endpoint || '').trim()) {
          reasons.push(intl.formatMessage({ id: 'pages.openclaw.userStats.reason.noEndpoint' }));
        }
        if (!isOkState(item.runtimeInsights?.statusState)) {
          reasons.push(`Runtime:${item.runtimeInsights?.statusState || '-'}`);
        }
        if (!isOkState(item.runtimeInsights?.gatewayState)) {
          reasons.push(`Gateway:${item.runtimeInsights?.gatewayState || '-'}`);
        }
        if (!isOkState(item.runtimeInsights?.securityState)) {
          reasons.push(`Security:${item.runtimeInsights?.securityState || '-'}`);
        }
        return {
          key: `${item.namespace || 'ns'}-${item.name || 'name'}`,
          namespace: item.namespace || '',
          name: item.name || '',
          displayName: item.displayName || item.name || '-',
          reasons,
        };
      })
      .filter((item) => item.reasons.length > 0)
      .slice(0, 8);
    return rows;
  }, [instanceRows, intl]);
  const healthScore = useMemo(() => {
    if (instanceRows.length === 0) {
      return 100;
    }
    const okWeight =
      runtimeSummary.statusOk + runtimeSummary.gatewayOk + runtimeSummary.securityOk + runningInstances;
    const fullWeight = instanceRows.length * 4;
    return Math.max(0, Math.min(100, Math.round((okWeight / fullWeight) * 100)));
  }, [
    instanceRows.length,
    runtimeSummary.statusOk,
    runtimeSummary.gatewayOk,
    runtimeSummary.securityOk,
    runningInstances,
  ]);
  const latestCollectedAt = useMemo(() => {
    const times = instanceRows
      .map((item) => String(item.runtimeInsights?.collectedAt || '').trim())
      .filter((x) => Boolean(x))
      .sort();
    return times.length > 0 ? times[times.length - 1] : '-';
  }, [instanceRows]);

  return (
    <PageContainer
      title={intl.formatMessage({ id: 'pages.openclaw.userStats.title' })}
      subTitle={intl.formatMessage({ id: 'pages.openclaw.userStats.subtitle' })}
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
          <Row gutter={[16, 16]} className={styles.statRow}>
            <Col xs={24} sm={24} md={24} lg={12} xl={12}>
              <Card title={intl.formatMessage({ id: 'pages.openclaw.userStats.posture.title' })}>
                <Row gutter={[12, 12]}>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.userStats.posture.health' })}
                      value={healthScore}
                      suffix="/100"
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.userStats.posture.attention' })}
                      value={interventionRows.length}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.userStats.posture.updatedAt' })}
                      value={latestCollectedAt === '-' ? '-' : 'OK'}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
            <Col xs={24} sm={24} md={24} lg={12} xl={12}>
              <Card title={intl.formatMessage({ id: 'pages.openclaw.userStats.usage.title' })}>
                <Row gutter={[12, 12]}>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.tokens.input' })}
                      value={toNumber(usageSummary.input).toFixed(0)}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.tokens.output' })}
                      value={toNumber(usageSummary.output).toFixed(0)}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.cost' })}
                      value={toNumber(usageSummary.cost).toFixed(4)}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
          </Row>

          <Row gutter={[16, 16]} className={styles.statRow}>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.kpi.instances' })}
                  value={toNumber(data.overview?.totalInstances)}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.kpi.accessible' })}
                  value={toNumber(data.overview?.accessibleInstances)}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.kpi.ready' })}
                  value={runningInstances}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.userStats.status.ok' })}
                  value={runtimeSummary.statusOk}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.userStats.gateway.ok' })}
                  value={runtimeSummary.gatewayOk}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={8} lg={4} xl={4}>
              <Card>
                <Statistic
                  title={intl.formatMessage({ id: 'pages.openclaw.userStats.security.ok' })}
                  value={runtimeSummary.securityOk}
                />
              </Card>
            </Col>
          </Row>

          <Card
            className={styles.tableCard}
            title={intl.formatMessage({ id: 'pages.openclaw.userStats.intervention.title' })}
            bodyStyle={{ paddingTop: 8, paddingBottom: 8 }}
          >
            {interventionRows.length === 0 ? (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={intl.formatMessage({ id: 'pages.openclaw.userStats.intervention.empty' })}
              />
            ) : (
              <Space direction="vertical" size={8} style={{ width: '100%' }}>
                {interventionRows.map((item) => {
                  const detailPath = `/openclaw/instance/${encodeURIComponent(
                    String(item.namespace || ''),
                  )}/${encodeURIComponent(String(item.name || ''))}`;
                  return (
                    <div key={item.key} className={styles.interventionRow}>
                      <div>
                        <strong>{item.displayName}</strong>
                        <div className={styles.providerMeta}>{item.reasons.join(' | ')}</div>
                      </div>
                      <Button size="small" icon={<EyeOutlined />} onClick={() => history.push(detailPath)}>
                        {intl.formatMessage({ id: 'pages.openclaw.action.view' })}
                      </Button>
                    </div>
                  );
                })}
              </Space>
            )}
          </Card>

          <Card
            className={styles.tableCard}
            title={intl.formatMessage({ id: 'pages.openclaw.user.table.title' })}
            bodyStyle={{ padding: 12 }}
          >
            {instanceRows.length === 0 ? (
              <Empty description={intl.formatMessage({ id: 'pages.openclaw.dashboard.empty.desc' })} />
            ) : (
              <Row gutter={[12, 12]}>
                {instanceRows.map((row) => {
                  const detailPath = `/openclaw/instance/${encodeURIComponent(
                    String(row.namespace || ''),
                  )}/${encodeURIComponent(String(row.name || ''))}`;
                  const runtime = row.runtimeInsights;
                  const runtimeMetaParts: string[] = [];
                  if (runtime?.runtimeVersion) {
                    runtimeMetaParts.push(`v:${runtime.runtimeVersion}`);
                  }
                  if (runtime?.openclawWorkspace) {
                    runtimeMetaParts.push(`ws:${runtime.openclawWorkspace}`);
                  }
                  if (runtime?.openclawPrimaryModel) {
                    runtimeMetaParts.push(`model:${runtime.openclawPrimaryModel}`);
                  }
                  return (
                    <Col xs={24} sm={12} xl={8} key={`${row.namespace || 'ns'}-${row.name || 'name'}`}>
                      <Card
                        size="small"
                        className={styles.instanceCompactCard}
                        title={row.displayName || row.name || '-'}
                        extra={
                          Number(row.readyPods || 0) > 0 && Boolean(String(row.endpoint || '').trim()) ? (
                            <Tag color="success">{intl.formatMessage({ id: 'pages.openclaw.value.yes' })}</Tag>
                          ) : (
                            <Tag>{intl.formatMessage({ id: 'pages.openclaw.value.no' })}</Tag>
                          )
                        }
                      >
                        <div className={styles.instanceCompactMeta}>{row.namespace || '-'}</div>
                        <div className={styles.instanceCompactPurpose}>{row.purpose || '-'}</div>
                        <Space wrap size={[6, 6]} className={styles.instanceCompactStates}>
                          <Tag color={stateTagColor(runtime?.statusState)}>S:{runtime?.statusState || '-'}</Tag>
                          <Tag color={stateTagColor(runtime?.gatewayState)}>G:{runtime?.gatewayState || '-'}</Tag>
                          <Tag color={stateTagColor(runtime?.securityState)}>A:{runtime?.securityState || '-'}</Tag>
                        </Space>
                        <div className={styles.providerMeta}>{runtimeMetaParts.join(' | ') || '-'}</div>
                        {runtime?.lastError ? (
                          <div className={styles.providerMeta}>{runtime.lastError}</div>
                        ) : null}
                        <div className={styles.instanceCompactFooter}>
                          <span className={styles.providerMeta}>{row.updatedAt || '-'}</span>
                          <Space size={8}>
                            <Button
                              size="small"
                              icon={<EyeOutlined />}
                              disabled={!row.namespace || !row.name}
                              onClick={() => history.push(detailPath)}
                            >
                              {intl.formatMessage({ id: 'pages.openclaw.action.view' })}
                            </Button>
                          </Space>
                        </div>
                      </Card>
                    </Col>
                  );
                })}
              </Row>
            )}
          </Card>
        </Spin>
      </div>
    </PageContainer>
  );
};

export default OpenClawUserStatsPage;
