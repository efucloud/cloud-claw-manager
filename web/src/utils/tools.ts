export const safeFormatJson = (
  jsonStr: string,
  space: number | string = 2
): string => {
  if (typeof jsonStr !== "string") {
    throw new TypeError("Input must be a string");
  }

  try {
    const parsed = JSON.parse(jsonStr.trim());
    return JSON.stringify(parsed, null, space);
  } catch (err) {
    console.error("Failed to parse JSON:", err);
    return jsonStr; // 或返回原字符串，或抛出错误
  }
};
