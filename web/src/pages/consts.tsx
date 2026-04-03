import { FormattedMessage } from 'react-intl';
export const BooleanValueTypeEnum = {
  false: {
    text: <FormattedMessage id="model.false" />,
    status: 'Error',
  },
  true: {
    text: <FormattedMessage id="model.true" />,
    status: 'Success',
  },
};
export const FieldValueTypeEnum = {
  dateTime: { text: <FormattedMessage id="model.field.valueType.dateTime" /> },
  string: { text: <FormattedMessage id="model.field.valueType.string" /> },
  number: { text: <FormattedMessage id="model.field.valueType.number" /> },
  float: { text: <FormattedMessage id="model.field.valueType.float" /> },
  bool: { text: <FormattedMessage id="model.field.valueType.bool" /> },
  json: { text: <FormattedMessage id="model.field.valueType.json" /> },
};

export const FieldValueTypeOptions = [
  {
    label: <FormattedMessage id="model.field.valueType.dateTime" />,
    value: 'dateTime',
  },
  {
    label: <FormattedMessage id="model.field.valueType.string" />,
    value: 'string',
  },
  {
    label: <FormattedMessage id="model.field.valueType.number" />,
    value: 'number',
  },

  {
    label: <FormattedMessage id="model.field.valueType.float" />,
    value: 'float',
  },
  {
    label: <FormattedMessage id="model.field.valueType.bool" />,
    value: 'bool',
  },
  {
    label: <FormattedMessage id="model.field.valueType.json" />,
    value: 'json',
  },
];
export const EsIndexDurationEnum = {
  hourly: {
    text: <FormattedMessage id="model.event_source.esIndexDuration.hourly" />,
  },
  daily: {
    text: <FormattedMessage id="model.event_source.esIndexDuration.daily" />,
  },
  weekly: {
    text: <FormattedMessage id="model.event_source.esIndexDuration.weekly" />,
  },
  monthly: {
    text: <FormattedMessage id="model.event_source.esIndexDuration.monthly" />,
  },
};

export const EsIndexDurationOptions = [
  {
    label: <FormattedMessage id="model.event_source.esIndexDuration.hourly" />,
    value: 'hourly',
  },
  {
    label: <FormattedMessage id="model.event_source.esIndexDuration.daily" />,
    value: 'daily',
  },

  {
    label: <FormattedMessage id="model.event_source.esIndexDuration.weekly" />,
    value: 'weekly',
  },
  {
    label: <FormattedMessage id="model.event_source.esIndexDuration.monthly" />,
    value: 'monthly',
  },
];
export const FignerprintTypeOptions = [
  {
    label: <FormattedMessage id="model.field.valueType.string" />,
    value: 'string',
  },
  {
    label: <FormattedMessage id="model.field.valueType.map" />,
    value: 'map',
  },
];
export const FignerprintValueEnum = {
  string: { text: <FormattedMessage id="model.field.valueType.string" /> },
  map: { text: <FormattedMessage id="model.field.valueType.map" /> },
};

export const RecentAlertOptions = [
  {
    label: <FormattedMessage id="model.event_alert.recent.1hour" />,
    value: '1hour',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.30minute" />,
    value: '30minute',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.1day" />,
    value: '1day',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.2day" />,
    value: '2day',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.1week" />,
    value: '1week',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.1month" />,
    value: '1month',
  },
  {
    label: <FormattedMessage id="model.event_alert.recent.halfyear" />,
    value: 'halfyear',
  },
];
