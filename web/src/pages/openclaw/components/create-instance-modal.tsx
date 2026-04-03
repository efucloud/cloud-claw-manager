import { openclawCreateInstance, openclawListTemplates } from '@/services/openclaw.api';
import { CheckCircleOutlined } from '@ant-design/icons';
import { useIntl } from '@umijs/max';
import { Button, Card, Col, Empty, Form, Input, InputNumber, Modal, Row, Select, Space, Spin, Steps, Switch, Typography, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import type { OpenClawTemplate } from '../types';
import styles from './create-instance-modal.less';

const { Text } = Typography;

type Props = {
  visible: boolean;
  onClose: () => void;
  onCreated: () => void;
};

const getSchemaProperties = (tpl?: OpenClawTemplate) => {
  const props = tpl?.schema?.properties;
  if (!props || typeof props !== 'object') return {} as Record<string, any>;
  return props as Record<string, any>;
};

const getRequiredSet = (tpl?: OpenClawTemplate) => {
  const required = tpl?.schema?.required;
  if (!Array.isArray(required)) return new Set<string>();
  return new Set(required.map((item) => String(item)));
};

const normalizeValuesBySchema = (values: Record<string, any>, schemaProperties: Record<string, any>) => {
  const out: Record<string, any> = {};
  for (const [key, propRaw] of Object.entries(schemaProperties)) {
    const prop = (propRaw || {}) as Record<string, any>;
    let value = values[key];
    if (value === undefined || value === null) {
      continue;
    }
    if (typeof value === 'string') {
      value = value.trim();
    }
    if (value === '' && prop.type !== 'string') {
      continue;
    }

    switch (String(prop.type || '').trim()) {
      case 'integer': {
        const n = Number(value);
        if (!Number.isFinite(n) || !Number.isInteger(n)) {
          return { errorKey: 'pages.openclaw.create.error.integer', field: String(prop.title || key) };
        }
        out[key] = n;
        break;
      }
      case 'number': {
        const n = Number(value);
        if (!Number.isFinite(n)) {
          return { errorKey: 'pages.openclaw.create.error.number', field: String(prop.title || key) };
        }
        out[key] = n;
        break;
      }
      case 'boolean': {
        out[key] = Boolean(value);
        break;
      }
      case 'object':
      case 'array': {
        if (typeof value === 'string') {
          if (!value) {
            continue;
          }
          try {
            out[key] = JSON.parse(value);
          } catch (_err) {
            return { errorKey: 'pages.openclaw.create.error.json', field: String(prop.title || key) };
          }
        } else {
          out[key] = value;
        }
        break;
      }
      default:
        out[key] = value;
        break;
    }
  }
  return { values: out };
};

const isSystemManagedField = (key: string) => {
  const name = String(key || '').trim().toLowerCase();
  return (
    name === 'ownerid' ||
    name === 'owner_id' ||
    name === 'namespace' ||
    name === 'previewendpoint' ||
    name === 'preview_endpoint' ||
    name === 'instancename' ||
    name === 'displayname' ||
    name === 'openclawname' ||
    name === 'purpose' ||
    name === 'openclawpurpose' ||
    name === 'gatewaytoken' ||
    name === 'openclawgatewaytoken' ||
    name === 'openclaw_gateway_token' ||
    name === 'modelapikey' ||
    name === 'model_api_key' ||
    name === 'openclawmodelapikey' ||
    name === 'openclaw_model_api_key' ||
    name === 'modelapikeysecretname' ||
    name === 'model_api_key_secret_name' ||
    name === 'modelapikeysecretkey' ||
    name === 'model_api_key_secret_key'
  );
};

const OpenClawCreateInstanceModal: React.FC<Props> = ({ visible, onClose, onCreated }) => {
  const intl = useIntl();
  const [templates, setTemplates] = useState<OpenClawTemplate[]>([]);
  const [templateLoading, setTemplateLoading] = useState(false);
  const [selectedTemplateName, setSelectedTemplateName] = useState('');
  const [step, setStep] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<Record<string, any>>();

  const selectedTemplate = useMemo(
    () => templates.find((item) => item.name === selectedTemplateName),
    [templates, selectedTemplateName],
  );

  const schemaProperties = useMemo(() => getSchemaProperties(selectedTemplate), [selectedTemplate]);
  const requiredSet = useMemo(() => getRequiredSet(selectedTemplate), [selectedTemplate]);

  const loadTemplates = useCallback(async () => {
    setTemplateLoading(true);
    try {
      const resp = await openclawListTemplates<any>();
      const source = resp as any;
      let rows: OpenClawTemplate[] = [];
      if (Array.isArray(source)) {
        rows = source;
      } else if (Array.isArray(source?.data)) {
        rows = source.data;
      } else if (Array.isArray(source?.data?.data)) {
        rows = source.data.data;
      } else if (Array.isArray(source?.result?.data)) {
        rows = source.result.data;
      }
      setTemplates(rows);
      setSelectedTemplateName((prev) => {
        if (prev && rows.some((item: OpenClawTemplate) => item.name === prev)) {
          return prev;
        }
        return rows[0]?.name || '';
      });
    } catch (_err) {
      message.error(intl.formatMessage({ id: 'pages.openclaw.create.template.load.failed' }));
      setTemplates([]);
      setSelectedTemplateName('');
    } finally {
      setTemplateLoading(false);
    }
  }, [intl]);

  useEffect(() => {
    if (!visible) {
      return;
    }
    setStep(0);
    void loadTemplates();
  }, [visible, loadTemplates]);

  useEffect(() => {
    if (!visible || step !== 1 || !selectedTemplate) {
      return;
    }
    form.resetFields();
    form.setFieldsValue(selectedTemplate.defaults || {});
  }, [visible, step, selectedTemplate, form]);

  const handleDeploy = useCallback(async () => {
    if (!selectedTemplate) {
      message.warning(intl.formatMessage({ id: 'pages.openclaw.create.template.select.required' }));
      return;
    }
    setSubmitting(true);
    try {
      const raw = await form.validateFields();
      const displayName = String(raw.displayName || '').trim();
      const purpose = String(raw.purpose || '').trim();
      const normalized = normalizeValuesBySchema(raw, schemaProperties);
      if ((normalized as any).errorKey) {
        message.error(
          intl.formatMessage(
            { id: (normalized as any).errorKey },
            { field: (normalized as any).field || '-' },
          ),
        );
        return;
      }
      const templateValues = (normalized as any).values || {};
      if (displayName) {
        templateValues.displayName = displayName;
        templateValues.openclawName = displayName;
      }
      if (purpose) {
        templateValues.purpose = purpose;
        templateValues.openclawPurpose = purpose;
      }
      if (!selectedTemplate.namespace) {
        message.error(intl.formatMessage({ id: 'pages.openclaw.create.namespace.missing' }));
        return;
      }
      await openclawCreateInstance({
        namespace: selectedTemplate.namespace,
        displayName,
        purpose,
        templateRef: selectedTemplate.name,
        templateValues,
      });
      message.success(intl.formatMessage({ id: 'pages.openclaw.create.deploy.success' }));
      onCreated();
      onClose();
      form.resetFields();
      setStep(0);
    } catch (err: any) {
      if (err?.errorFields) {
        return;
      }
      message.error(err?.message || intl.formatMessage({ id: 'pages.openclaw.create.deploy.failed' }));
    } finally {
      setSubmitting(false);
    }
  }, [selectedTemplate, schemaProperties, form, onCreated, onClose, intl]);

  const renderSchemaFieldWithIntl = (key: string, prop: Record<string, any>) => {
    if (isSystemManagedField(key)) {
      return null;
    }
    const title = String(prop.title || key);
    const description = String(prop.description || '');
    const typeName = String(prop.type || '').trim();
    const required = requiredSet.has(key);
    const enumValues = Array.isArray(prop.enum) ? prop.enum : null;
    const fieldFormat = String(prop.format || '').trim().toLowerCase();
    const isSensitive =
      fieldFormat === 'password' ||
      fieldFormat === 'secret' ||
      String(key).toLowerCase().includes('token');
    const requiredRules = required
      ? [{ required: true, message: intl.formatMessage({ id: 'pages.openclaw.create.form.required' }, { field: title }) }]
      : undefined;

    if (enumValues) {
      return (
        <Form.Item key={key} name={key} label={title} required={required} rules={requiredRules} extra={description || undefined}>
          <Select
            options={enumValues.map((item: any) => ({ label: String(item), value: item }))}
            placeholder={intl.formatMessage({ id: 'pages.openclaw.create.form.placeholder.select' }, { field: title })}
            allowClear={!required}
          />
        </Form.Item>
      );
    }

    if (typeName === 'boolean') {
      return (
        <Form.Item key={key} name={key} label={title} valuePropName="checked" extra={description || undefined}>
          <Switch />
        </Form.Item>
      );
    }

    if (typeName === 'integer' || typeName === 'number') {
      return (
        <Form.Item key={key} name={key} label={title} required={required} rules={requiredRules} extra={description || undefined}>
          <InputNumber
            style={{ width: '100%' }}
            placeholder={intl.formatMessage({ id: 'pages.openclaw.create.form.placeholder.input' }, { field: title })}
          />
        </Form.Item>
      );
    }

    if (typeName === 'object' || typeName === 'array') {
      return (
        <Form.Item
          key={key}
          name={key}
          label={title}
          required={required}
          rules={requiredRules}
          extra={description || intl.formatMessage({ id: 'pages.openclaw.create.form.hint.json' })}
        >
          <Input.TextArea
            autoSize={{ minRows: 3, maxRows: 6 }}
            placeholder={intl.formatMessage({ id: 'pages.openclaw.create.form.placeholder.json' }, { field: title })}
          />
        </Form.Item>
      );
    }

    return (
      <Form.Item key={key} name={key} label={title} required={required} rules={requiredRules} extra={description || undefined}>
        {isSensitive ? (
          <Input.Password placeholder={intl.formatMessage({ id: 'pages.openclaw.create.form.placeholder.input' }, { field: title })} />
        ) : (
          <Input placeholder={intl.formatMessage({ id: 'pages.openclaw.create.form.placeholder.input' }, { field: title })} />
        )}
      </Form.Item>
    );
  };

  return (
    <Modal
      open={visible}
      title={intl.formatMessage({ id: 'pages.openclaw.create.title' })}
      width={980}
      destroyOnClose
      maskClosable={false}
      onCancel={onClose}
      footer={
        step === 0
          ? [
              <Button key="cancel" onClick={onClose}>{intl.formatMessage({ id: 'pages.operation.cancel' })}</Button>,
              <Button key="next" type="primary" disabled={!selectedTemplateName} onClick={() => setStep(1)}>
                {intl.formatMessage({ id: 'pages.operation.next' })}
              </Button>,
            ]
          : [
              <Button key="cancel" onClick={onClose}>{intl.formatMessage({ id: 'pages.operation.cancel' })}</Button>,
              <Button key="prev" onClick={() => setStep(0)}>{intl.formatMessage({ id: 'pages.operation.prev' })}</Button>,
              <Button key="deploy" type="primary" loading={submitting} onClick={() => void handleDeploy()}>
                {intl.formatMessage({ id: 'pages.operation.deploy' })}
              </Button>,
            ]
      }
    >
      <Steps
        size="small"
        current={step}
        className={styles.steps}
        items={[
          { title: intl.formatMessage({ id: 'pages.openclaw.create.step.selectTemplate' }) },
          { title: intl.formatMessage({ id: 'pages.openclaw.create.step.config' }) },
        ]}
      />

      {step === 0 ? (
        <Spin spinning={templateLoading}>
          {templates.length === 0 ? (
            <Empty description={intl.formatMessage({ id: 'pages.openclaw.create.template.empty' })} />
          ) : (
            <Row gutter={[12, 12]}>
              {templates.map((item) => {
                const selected = item.name === selectedTemplateName;
                return (
                  <Col xs={24} md={12} key={item.name} className={styles.templateCol}>
                    <Card
                      hoverable
                      className={`${styles.templateCard} ${selected ? styles.templateCardSelected : ''}`}
                      onClick={() => setSelectedTemplateName(item.name)}
                    >
                      <Space direction="vertical" size={4}>
                        <Space size={6}>
                          {selected ? <CheckCircleOutlined className={styles.selectedIcon} /> : null}
                          <Text strong>{item.name}</Text>
                        </Space>
                        <Text type="secondary">{item.description || intl.formatMessage({ id: 'pages.openclaw.create.template.noDesc' })}</Text>
                      </Space>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          )}
        </Spin>
      ) : (
        <Form form={form} layout="vertical" className={styles.form}>
          <Form.Item
            name="displayName"
            label={intl.formatMessage({ id: 'pages.openclaw.create.field.displayName' })}
            rules={[
              {
                required: true,
                message: intl.formatMessage(
                  { id: 'pages.openclaw.create.form.required' },
                  { field: intl.formatMessage({ id: 'pages.openclaw.create.field.displayName' }) },
                ),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage(
                { id: 'pages.openclaw.create.form.placeholder.input' },
                { field: intl.formatMessage({ id: 'pages.openclaw.create.field.displayName' }) },
              )}
            />
          </Form.Item>
          <Form.Item
            name="purpose"
            label={intl.formatMessage({ id: 'pages.openclaw.create.field.purpose' })}
            rules={[
              {
                required: true,
                message: intl.formatMessage(
                  { id: 'pages.openclaw.create.form.required' },
                  { field: intl.formatMessage({ id: 'pages.openclaw.create.field.purpose' }) },
                ),
              },
            ]}
          >
            <Input.TextArea
              autoSize={{ minRows: 2, maxRows: 4 }}
              placeholder={intl.formatMessage(
                { id: 'pages.openclaw.create.form.placeholder.input' },
                { field: intl.formatMessage({ id: 'pages.openclaw.create.field.purpose' }) },
              )}
            />
          </Form.Item>
          {Object.entries(schemaProperties).map(([key, prop]) => renderSchemaFieldWithIntl(key, (prop || {}) as Record<string, any>))}
        </Form>
      )}
    </Modal>
  );
};

export default OpenClawCreateInstanceModal;
