// hooks/useAiCopilot.ts
import type {
  ChatRequest,
  ChatStopMeta,
  StreamEvent,
  ToolChain,
} from "@/services/ai_copilot";
import { getI18nLanguage, getToken } from "@/utils/global";
import { parseNDJSONStream } from "@/utils/ndjson-stream";
import { getLocale, useIntl } from "@umijs/max";
import { message } from "antd";
import { useCallback, useEffect, useRef, useState } from "react";

export interface UseAiCopilotOptions {
  cluster: string;
  namespace?: string;
}
export interface ExtendedStreamEvent extends StreamEvent {
  question: string;
  tools?: Array<{
    name: string;
    description?: string;
    status: "success" | "error";
  }>;
}

export type PlanStepStatus = "pending" | "executing" | "success" | "error";

export interface PlanStepState {
  id?: string;
  title?: string;
  tool?: string;
  reason?: string;
  status: PlanStepStatus;
  detail?: string;
  callId?: string;
  turn?: number;
}

export interface PlanRunState {
  requestId?: string;
  goal?: string;
  mode?: "plan" | "agent";
  steps: PlanStepState[];
}

type PlanEventData = {
  goal?: string;
  steps?: Array<{
    id?: string;
    title?: string;
    tool?: string;
    reason?: string;
  }>;
};

const toPlanStatus = (status?: string): PlanStepStatus => {
  if (status === "executing") {
    return "executing";
  }
  if (status === "success") {
    return "success";
  }
  if (status === "error") {
    return "error";
  }
  return "pending";
};

const parsePlanData = (data: unknown): PlanEventData | undefined => {
  if (!data) return undefined;
  if (typeof data === "string") {
    try {
      return JSON.parse(data) as PlanEventData;
    } catch {
      return undefined;
    }
  }
  return data as PlanEventData;
};

const buildPlanStateFromData = (
  requestId: string,
  data: unknown
): PlanRunState | undefined => {
  const parsed = parsePlanData(data);
  if (!parsed || !Array.isArray(parsed.steps)) {
    return undefined;
  }
  return {
    requestId,
    goal: parsed.goal || "",
    mode: "plan",
    steps: parsed.steps.map((step) => ({
      id: step.id,
      title: step.title,
      tool: step.tool,
      reason: step.reason,
      status: "pending" as PlanStepStatus,
    })),
  };
};

const resolveRequestLocale = () => {
  let currentLocale = getLocale();
  if (!currentLocale) {
    currentLocale = getI18nLanguage();
  }
  if (!currentLocale && typeof window !== "undefined") {
    currentLocale = localStorage.getItem("umi_locale") || "zh-CN";
  }
  return currentLocale || "zh-CN";
};

const normalizeDetailValue = (value: unknown): string | undefined => {
  if (value === null || value === undefined) {
    return undefined;
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "object") {
    try {
      return JSON.stringify(value);
    } catch {
      return String(value);
    }
  }
  return String(value);
};
/**
 * AI 流式对话 Hook（基于 Streamable HTTP + NDJSON）
 * 支持动态传入 ChatRequest，实现多轮、多上下文对话
 */
export function useAiCopilot(options: UseAiCopilotOptions) {
  const intl = useIntl();
  const [loading, setLoading] = useState(false);
  const [latestContent, setLatestContent] = useState<ExtendedStreamEvent>();
  const [aiChatUsage, setAiChatUsage] = useState<ChatStopMeta>();
  const [errors, setErrors] = useState<string[]>([]);
  const [sessionId, setSessionId] = useState<string | undefined>(undefined);
  const abortControllerRef = useRef<AbortController | null>(null);
  const [usedTool, setUsedTool] = useState<string>("");
  const [toolChains, setToolChains] = useState<ToolChain[]>([]);
  const [planRuns, setPlanRuns] = useState<Record<string, PlanRunState>>({});
  const isMountedRef = useRef(true);
  const { cluster, namespace } = options;
  const orgToken = getToken();

  //   发送 AI 请求（接收完整的 ChatRequest）
  const sendMessage = useCallback(
    async (request: ChatRequest) => {
      // 重置状态
      setLatestContent(undefined);
      setAiChatUsage(undefined);
      setErrors([]);
      setSessionId(undefined);
      // 取消上一次请求
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      const controller = new AbortController();
      abortControllerRef.current = controller;
      setLoading(true);
      try {
        const locale = resolveRequestLocale();
        // 构建 URL：仅当 namespace 有效时加入路径段
        const baseUrl = namespace
          ? `/api/stream/cluster/${cluster}/namespace/${namespace}`
          : `/api/stream/cluster/${cluster}`;
        const response = await fetch(baseUrl, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${orgToken?.access_token || ""}`,
            "X-Locale": locale,
          },
          body: JSON.stringify(request),
          signal: controller.signal,
        });

        if (!response.ok) {
          const errorText = await response.text().catch(() => "Unknown error");
          throw new Error(`HTTP ${response.status}: ${errorText}`);
        }

        // 解析 NDJSON 流
        await parseNDJSONStream(
          response,
          (event: StreamEvent) => {
            if (!isMountedRef.current || controller.signal.aborted) return;

            if (event.error) {
              if (event.error !== "") {
                const errorText = event.error || event.content || "";
                setErrors((prev) => [...prev, event.error || ""]);
                setLatestContent((prev: ExtendedStreamEvent | undefined) => {
                  if (!prev) {
                    return {
                      ...event,
                      requestId: event.requestId || request.requestId,
                      question: request.message || "",
                      content: errorText,
                    };
                  } else {
                    return {
                      ...prev,
                      content: (prev.content || "") + errorText,
                    };
                  }
                });
              }
              return;
            }
            if (event.type === "token") {
              setLatestContent((prev: ExtendedStreamEvent | undefined) => {
                if (!prev) {
                  return { ...event, question: request.message || "" };
                } else {
                  return {
                    ...prev,
                    content: (prev.content || "") + (event.content || ""),
                  };
                }
              });
            } else if (event.type === "done") {
              if (event.meta) {
                try {
                  const data = JSON.parse(
                    JSON.stringify(event.meta, null, 2)
                  ) as ChatStopMeta;
                  setAiChatUsage(data);
                } catch (error) {
                  console.log(error);
                }
              }

              const doneMeta = event.meta as
                | { mode?: string; auto_executed?: boolean }
                | undefined;
              const doneRequestId = event.requestId || request.requestId || "";
              if (doneRequestId) {
                setPlanRuns((prev) => {
                  const existed = prev[doneRequestId];
                  if (!existed) {
                    return prev;
                  }
                  const mergedSteps = existed.steps.map((step) => {
                    if (
                      doneMeta?.mode === "plan" &&
                      doneMeta?.auto_executed &&
                      step.status === "pending" &&
                      !step.tool
                    ) {
                      return { ...step, status: "success" as PlanStepStatus };
                    }
                    if (step.status === "executing") {
                      return { ...step, status: "success" as PlanStepStatus };
                    }
                    return step;
                  });
                  return {
                    ...prev,
                    [doneRequestId]: {
                      ...existed,
                      mode:
                        (doneMeta?.mode as "plan" | "agent" | undefined) ||
                        existed.mode,
                      steps: mergedSteps,
                    },
                  };
                });
              }
              setLoading(false);
            } else if (event.type === "tool") {
              const toolMeta = event.meta as
                | (ToolChain & {
                    callId?: string;
                    stepId?: string;
                    title?: string;
                    description?: string;
                  })
                | undefined;
              const toolData = event.data as
                | {
                    mode?: string;
                    status?: string;
                    arguments?: string;
                    result?: string;
                    error?: string;
                    callId?: string;
                    stepId?: string;
                    title?: string;
                    tool?: string;
                    turn?: number;
                    reason?: string;
                  }
                | undefined;
              const chainRequestId =
                toolMeta?.requestId || event.requestId || request.requestId;
              const chainCallId = toolData?.callId || toolMeta?.callId || "";
              const chainStepId = toolData?.stepId || toolMeta?.stepId || "";
              const chainTitle =
                toolData?.title || toolMeta?.title || event.tool || "";
              const chainStatus =
                toolMeta?.status || (toolData?.status as string) || "executing";
              const chainDescription =
                toolMeta?.description ||
                normalizeDetailValue(toolData?.error) ||
                normalizeDetailValue(toolData?.result) ||
                normalizeDetailValue(toolData?.arguments) ||
                "";

              const toolChain: ToolChain = {
                title: chainTitle,
                description: chainDescription,
                status: chainStatus,
                requestId: chainRequestId,
              };

              setToolChains((prev) => {
                const index = prev.findIndex((tc) => {
                  const current = tc as ToolChain & {
                    callId?: string;
                    stepId?: string;
                  };
                  if (current.requestId !== chainRequestId) {
                    return false;
                  }
                  if (chainCallId && current.callId === chainCallId) {
                    return true;
                  }
                  if (chainStepId && current.stepId === chainStepId) {
                    return true;
                  }
                  return current.title === chainTitle;
                });

                const nextTool = {
                  ...(index >= 0 ? (prev[index] as ToolChain) : {}),
                  ...toolChain,
                  callId: chainCallId,
                  stepId: chainStepId,
                } as ToolChain;

                if (index >= 0) {
                  const updated = [...prev];
                  updated[index] = nextTool;
                  return updated;
                } else {
                  return [...prev, nextTool];
                }
              });

              const toolRequestId = event.requestId || request.requestId || "";
              if (
                toolRequestId &&
                (toolData?.stepId || toolData?.title || toolData?.callId)
              ) {
                setPlanRuns((prev) => {
                  const existed = prev[toolRequestId] || {
                    requestId: toolRequestId,
                    goal: "",
                    mode:
                      (toolData?.mode as "plan" | "agent" | undefined) ||
                      undefined,
                    steps: [],
                  };
                  const steps = [...existed.steps];
                  const stepId = toolData?.stepId || toolData?.callId || "";
                  const stepTitle = toolData?.title || event.tool || "";
                  const stepStatus = toPlanStatus(toolData?.status);
                  const stepDetail =
                    normalizeDetailValue(toolData?.error) ||
                    normalizeDetailValue(toolData?.result) ||
                    (stepStatus === "executing"
                      ? normalizeDetailValue(toolData?.arguments)
                      : "");

                  let index = steps.findIndex((step) =>
                    stepId
                      ? step.id === stepId
                      : step.title === stepTitle && step.tool === event.tool
                  );
                  if (index < 0) {
                    steps.push({
                      id: stepId,
                      title: stepTitle,
                      tool: event.tool,
                      reason: toolData?.reason,
                      status: stepStatus,
                      detail: stepDetail,
                      callId: toolData?.callId,
                      turn: toolData?.turn,
                    });
                    index = steps.length - 1;
                  } else {
                    steps[index] = {
                      ...steps[index],
                      id: steps[index].id || stepId,
                      title: steps[index].title || stepTitle,
                      tool: steps[index].tool || event.tool,
                      reason: steps[index].reason || toolData?.reason,
                      status: stepStatus,
                      detail: stepDetail || steps[index].detail,
                      callId: steps[index].callId || toolData?.callId,
                      turn: steps[index].turn || toolData?.turn,
                    };
                  }

                  return {
                    ...prev,
                    [toolRequestId]: {
                      ...existed,
                      mode:
                        (toolData?.mode as "plan" | "agent" | undefined) ||
                        existed.mode,
                      goal:
                        existed.goal ||
                        ((toolData?.mode || "").toLowerCase() === "agent"
                          ? "Agent 执行轨迹"
                          : ""),
                      steps,
                    },
                  };
                });
              }
            } else if (event.type === "skill") {
              const skillId =
                event.content || (event.meta as { id?: string })?.id || "";
              setUsedTool(skillId);
            } else if (event.type === "plan") {
              const planRequestId = event.requestId || request.requestId || "";
              if (planRequestId) {
                const planState = buildPlanStateFromData(
                  planRequestId,
                  event.data
                );
                if (planState) {
                  setPlanRuns((prev) => ({
                    ...prev,
                    [planRequestId]: {
                      ...planState,
                      mode: "plan",
                    },
                  }));
                }
              }
              const planContent =
                typeof event.data === "string"
                  ? event.data
                  : `\`\`\`json\n${JSON.stringify(
                      event.data || {},
                      null,
                      2
                    )}\n\`\`\``;
              setLatestContent((prev: ExtendedStreamEvent | undefined) => {
                if (!prev) {
                  return {
                    ...event,
                    question: request.message || "",
                    content: planContent,
                  };
                } else {
                  return {
                    ...prev,
                    content:
                      (prev.content ? `${prev.content}\n\n` : "") + planContent,
                  };
                }
              });
            } else if (event.type === "error") {
              const errorText = event.error || event.content || "";
              setLatestContent((prev: ExtendedStreamEvent | undefined) => {
                if (!prev) {
                  return {
                    ...event,
                    requestId: event.requestId || request.requestId,
                    question: request.message || "",
                    content: errorText,
                  };
                } else {
                  return {
                    ...prev,
                    content: (prev.content || "") + errorText,
                  };
                }
              });
            }
            if (event.sessionId && !sessionId) {
              setSessionId(event.sessionId);
            }
          },
          (err) => {
            if (!isMountedRef.current || controller.signal.aborted) return;
            console.error("AI stream error:", err);
            setErrors((prev) => [...prev, err.message]);
            setLatestContent((prev: ExtendedStreamEvent | undefined) => {
              if (prev) {
                return {
                  ...prev,
                  content: (prev.content || "") + err.message || "",
                };
              } else {
                return {
                  question: request.message || "",
                  content: err.message || "",
                };
              }
            });
            message.error(
              `${intl.formatMessage({ id: "copilot.stream.error" })}: ${
                err.message
              }`
            );
          }
        );

        if (isMountedRef.current && !controller.signal.aborted) {
          setLoading(false);
        }
      } catch (err: any) {
        if (controller.signal.aborted) return;
        console.error("Failed to send AI request:", err);
        setErrors((prev) => [...prev, err.message]);
        message.error(
          `${intl.formatMessage({ id: "copilot.send.failed" })}: ${err.message}`
        );
        setLoading(false);
      }
    },
    [cluster, orgToken?.access_token]
  );

  /**
   * 取消当前请求
   */
  const cancelRequest = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      setLoading(false);
    }
  }, [latestContent]);

  // 清理副作用
  useEffect(() => {
    return () => {
      isMountedRef.current = false;
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return {
    loading,
    latestContent,
    aiChatUsage,
    errors,
    usedTool,
    sessionId,
    sendMessage,
    cancelRequest,
    toolChains,
    planRuns,
  };
}
