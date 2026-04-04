import { openclawDashboard } from '@/services/openclaw.api';
import { EyeOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history, useIntl } from '@umijs/max';
import { Button, Card, Col, Empty, Row, Space, Spin, Statistic, Table, Tag, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { normalizeProvider, normalizeUrlHost, toNumber } from './format';
import type { DashboardResponse } from './types';
import styles from './dashboard.less';

const OpenClawAdminDashboardPage: React.FC = () => {
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
  const totalInstances =
    data.overview?.totalInstances === undefined
      ? instanceRows.length
      : toNumber(data.overview?.totalInstances);
  const accessibleInstances =
    data.overview?.accessibleInstances === undefined
      ? instanceRows.filter((row) => Boolean(row.accessible)).length
      : toNumber(data.overview?.accessibleInstances);
  const runningInstances = useMemo(
    () =>
      instanceRows.filter(
        (row) => Number(row.readyPods || 0) > 0 && Boolean(String(row.endpoint || '').trim()),
      ).length,
    [instanceRows],
  );
  const providerTags = useMemo(() => {
    const breakdown = new Map<string, number>();
    const providers = data.providers || [];
    if (providers.length > 0) {
      providers.forEach((item) => {
        const key = normalizeProvider(item.provider) || 'unknown';
        breakdown.set(key, toNumber(breakdown.get(key)) + toNumber(item.requests));
      });
    } else {
      instanceRows.forEach((item) => {
        const key = normalizeProvider(item.provider) || 'unknown';
        breakdown.set(key, toNumber(breakdown.get(key)) + 1);
      });
    }
    return Array.from(breakdown.entries())
      .map(([provider, requests]) => ({ provider, requests }))
      .sort((a, b) => b.requests - a.requests)
      .slice(0, 8);
  }, [data.providers, instanceRows]);

  return (
    <PageContainer
      title={intl.formatMessage({ id: 'pages.openclaw.admin.title' })}
    >
      <div className={styles.wrapper}>
        <Spin spinning={loading}>
          {instanceRows.length === 0 ? (
            <div className={styles.emptyWrap}>
              <Empty description={intl.formatMessage({ id: 'pages.openclaw.dashboard.empty.desc' })} />
            </div>
          ) : (
            <>
              <Row gutter={[16, 16]} className={styles.statRow}>
                <Col xs={12} sm={12} md={8} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.instances' })}
                      value={totalInstances}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={12} md={8} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.accessible' })}
                      value={accessibleInstances}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={12} md={8} lg={6} xl={6}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.ready' })}
                      value={runningInstances}
                    />
                  </Card>
                </Col>
              </Row>

              <Card size="small" title={intl.formatMessage({ id: 'pages.openclaw.providers' })} className={styles.providerCard}>
                <Space wrap>
                  {providerTags.length === 0 ? (
                    <Tag>{intl.formatMessage({ id: 'pages.openclaw.value.unknown' })}</Tag>
                  ) : (
                    providerTags.map((item) => (
                      <Tag key={item.provider}>
                        {normalizeProvider(item.provider) ||
                          intl.formatMessage({ id: 'pages.openclaw.value.unknown' })}
                        : {toNumber(item.requests).toFixed(0)}
                      </Tag>
                    ))
                  )}
                </Space>
              </Card>

              <Card className={styles.tableCard} bodyStyle={{ padding: 0 }}>
                <Table
                  rowKey={(row) => `${row.namespace || 'ns'}-${row.name || 'name'}`}
                  dataSource={instanceRows}
                  pagination={{ pageSize: 10 }}
                  columns={[
                    {
                      title: intl.formatMessage({ id: 'pages.dashboard.field.namespace' }),
                      dataIndex: 'namespace',
                      width: 140,
                    },
                    {
                      title: intl.formatMessage({ id: 'pages.openclaw.field.displayName' }),
                      render: (_, row) => row.displayName || row.name || '-',
                      width: 180,
                    },
                    {
                      title: intl.formatMessage({ id: 'pages.openclaw.field.ownerUser' }),
                      render: (_, row) => row.ownerUsername || row.ownerEmail || row.ownerId || '-',
                    },
                    {
                      title: intl.formatMessage({ id: 'pages.openclaw.field.provider' }),
                      width: 260,
                      render: (_, row) => {
                        const provider =
                          normalizeProvider(row.provider) ||
                          intl.formatMessage({ id: 'pages.openclaw.value.unknown' });
                        const detailParts: string[] = [];
                        if (row.providerPrimary) {
                          detailParts.push(`primary: ${row.providerPrimary}`);
                        }
                        if ((row.providerModelIds || []).length > 0) {
                          detailParts.push(`models: ${(row.providerModelIds || []).join(', ')}`);
                        }
                        if (row.providerApiType) {
                          detailParts.push(`api: ${row.providerApiType}`);
                        }
                        const host = normalizeUrlHost(row.providerBaseUrl);
                        if (host) {
                          detailParts.push(`base: ${host}`);
                        }
                        return (
                          <div>
                            <div>{provider}</div>
                            {detailParts.length > 0 && (
                              <div className={styles.providerMeta}>{detailParts.join(' | ')}</div>
                            )}
                          </div>
                        );
                      },
                    },
                    {
                      title: intl.formatMessage({ id: 'pages.dashboard.state.ready' }),
                      width: 100,
                      render: (_, row) =>
                        (Number(row.readyPods || 0) > 0 && Boolean(String(row.endpoint || '').trim())) ? (
                          <Tag color="success">{intl.formatMessage({ id: 'pages.openclaw.value.yes' })}</Tag>
                        ) : (
                          <Tag>{intl.formatMessage({ id: 'pages.openclaw.value.no' })}</Tag>
                        ),
                    },
                    {
                      title: intl.formatMessage({ id: 'pages.operation' }),
                      width: 120,
                      render: (_, row) => {
                        const detailPath = `/openclaw/instance/${encodeURIComponent(
                          String(row.namespace || ''),
                        )}/${encodeURIComponent(String(row.name || ''))}`;
                        return (
                          <Button
                            size="small"
                            icon={<EyeOutlined />}
                            disabled={!row.namespace || !row.name}
                            onClick={() => history.push(detailPath)}
                          >
                            {intl.formatMessage({ id: 'pages.openclaw.action.view' })}
                          </Button>
                        );
                      },
                    },
                  ]}
                />
              </Card>
            </>
          )}
        </Spin>
      </div>
    </PageContainer>
  );
};

export default OpenClawAdminDashboardPage;
