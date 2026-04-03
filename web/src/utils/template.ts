export const getTemplateInfo = (
  template: string
): { apiVersion: string; kind: string; name: string } => {
  const lines = template.split("\n");
  let apiVersion = "";
  let kind = "";
  let name = "";
  for (const line of lines) {
    if (line.startsWith("apiVersion:")) {
      apiVersion = line.substring("apiVersion:".length).trim();
    } else if (line.startsWith("kind:")) {
      kind = line.substring("kind:".length).trim();
    } else if (line.startsWith("  name:")) {
      name = line.substring("  name:".length).trim();
    }

    // 可选：提前退出（如果三个都找到了）
    if (apiVersion && kind && name) {
      break;
    }
  }
  return { apiVersion, kind, name };
};
export const formatJsonString = (jsonStr: string): string => {
  if (!jsonStr || typeof jsonStr !== "string") {
    return "{}";
  }

  try {
    // 解析后再格式化，确保语法正确且带缩进
    const obj = JSON.parse(jsonStr);
    return JSON.stringify(obj, null, 2);
  } catch (error) {
    // 如果不是合法 JSON，直接返回原字符串（避免白屏）
    console.warn("Invalid JSON string:", jsonStr);
    return jsonStr;
  }
};

export const spaceToISO = (timeStr: string): string => {
  // 输入: "2026-01-09 00:00:08"
  // 输出: "2026-01-09T00:00:08Z" （假设是 UTC）
  return timeStr.replace(" ", "T") + "Z";
};
