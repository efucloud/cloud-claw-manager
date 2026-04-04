import { request } from '@umijs/max';

export type OpenClawTemplateUpsertRequest = {
  name?: string;
  description?: string;
  schema: Record<string, any>;
  defaults?: Record<string, any>;
  templates: string;
};

export type OpenClawInstanceCreateRequest = {
  namespace: string;
  name?: string;
  displayName?: string;
  purpose?: string;
  visibilityUsers?: string[];
  templateRef?: string;
  templateValues?: Record<string, any>;
  runtime?: Record<string, any>;
};

export async function openclawDashboard<T>(
  params?: {
    namespace?: string;
    includeGatewayToken?: boolean;
  },
  options?: { [key: string]: any },
) {
  return request<T>(`/api/v1/openclaw/dashboard`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    params,
    ...(options || {}),
  });
}

export async function openclawListTemplates<T>(options?: { [key: string]: any }) {
  return request<T>(`/api/v1/openclaw/templates`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}

export async function openclawGetTemplate<T>(name: string, options?: { [key: string]: any }) {
  return request<T>(`/api/v1/openclaw/templates/${name}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}

export async function openclawCreateTemplate<T>(
  body: OpenClawTemplateUpsertRequest,
  options?: { [key: string]: any },
) {
  return request<T>(`/api/v1/openclaw/templates`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

export async function openclawUpdateTemplate<T>(
  name: string,
  body: OpenClawTemplateUpsertRequest,
  options?: { [key: string]: any },
) {
  return request<T>(`/api/v1/openclaw/templates/${name}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

export async function openclawCreateInstance<T>(
  body: OpenClawInstanceCreateRequest,
  options?: { [key: string]: any },
) {
  return request<T>(`/api/v1/openclaw/instances`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

export async function openclawListInstances<T>(
  params?: {
    namespace?: string;
  },
  options?: { [key: string]: any },
) {
  return request<T>(`/api/v1/openclaw/instances`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    params,
    ...(options || {}),
  });
}

export async function openclawGetInstance<T>(
  namespace: string,
  name: string,
  options?: { [key: string]: any },
) {
  const ns = encodeURIComponent(namespace);
  const nm = encodeURIComponent(name);
  return request<T>(`/api/v1/openclaw/instances/${ns}/${nm}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
