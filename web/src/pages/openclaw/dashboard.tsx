import { openclawDashboard } from '@/services/openclaw.api';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history, useIntl, useLocation } from '@umijs/max';
import { Button, Card, Col, Empty, Row, Spin, Statistic, message } from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import OpenClawCreateInstanceModal from './components/create-instance-modal';
import OpenClawPreviewGrid from './components/preview-grid';
import { toNumber } from './format';
import type { DashboardResponse } from './types';
import styles from './dashboard.less';

const OpenClawDashboardPage: React.FC = () => {
  const intl = useIntl();
  const location = useLocation();
  const [loading, setLoading] = useState<boolean>(false);
  const [data, setData] = useState<DashboardResponse>({});
  const [createVisible, setCreateVisible] = useState<boolean>(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await openclawDashboard<DashboardResponse>();
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

  useEffect(() => {
    const params = new URLSearchParams(location.search || '');
    if (params.get('openclawCreate') === '1') {
      setCreateVisible(true);
    }
  }, [location.search]);

  const closeCreateModal = useCallback(() => {
    setCreateVisible(false);
    const params = new URLSearchParams(location.search || '');
    if (params.has('openclawCreate')) {
      params.delete('openclawCreate');
      const qs = params.toString();
      history.replace(`${location.pathname}${qs ? `?${qs}` : ''}`);
    }
  }, [location.pathname, location.search]);

  const handleCreated = useCallback(() => {
    void loadData();
  }, [loadData]);

  const instanceRows = data.instances || [];

  return (
    <PageContainer
      title={false}
      extra={[
        <Button key="refresh" icon={<ReloadOutlined />} loading={loading} onClick={() => void loadData()}>
          {intl.formatMessage({ id: 'pages.operation.refresh' })}
        </Button>,
        <Button key="user-stats" onClick={() => history.push('/dashboard/stats')}>
          {intl.formatMessage({ id: 'pages.openclaw.action.userStats' })}
        </Button>,
        <Button key="create" type="primary" icon={<PlusOutlined />} onClick={() => setCreateVisible(true)}>
          {intl.formatMessage({ id: 'pages.openclaw.action.create' })}
        </Button>,
      ]}
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
                <Col xs={12} sm={8} md={8} lg={8} xl={8}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.instances' })}
                      value={toNumber(data.overview?.totalInstances)}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={8} md={8} lg={8} xl={8}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.accessible' })}
                      value={toNumber(data.overview?.accessibleInstances)}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={8} md={8} lg={8} xl={8}>
                  <Card>
                    <Statistic
                      title={intl.formatMessage({ id: 'pages.openclaw.kpi.ready' })}
                      value={toNumber(data.overview?.readyPods)}
                    />
                  </Card>
                </Col>
              </Row>
              <OpenClawPreviewGrid instances={instanceRows} />
            </>
          )}
        </Spin>
      </div>

      <OpenClawCreateInstanceModal visible={createVisible} onClose={closeCreateModal} onCreated={handleCreated} />
    </PageContainer>
  );
};

export default OpenClawDashboardPage;
