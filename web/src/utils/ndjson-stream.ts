/**
 * 解析 NDJSON (Newline Delimited JSON) 流
 * @param response fetch 返回的 Response 对象
 * @param onMessage 每收到一行 JSON 的回调
 * @param onError 错误回调
 */
export async function parseNDJSONStream(
  response: Response,
  onMessage: (data: any) => void,
  onError?: (error: Error) => void
) {
  if (!response.body) {
    onError?.(new Error("Response body is null"));
    return;
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      // 按换行符分割
      const lines = buffer.split("\n");
      buffer = lines.pop() || ""; // 保留不完整的最后一行

      for (const line of lines) {
        if (line.trim() === "") continue; // 跳过空行
        try {
          const json = JSON.parse(line);
          onMessage(json);
        } catch (e) {
          console.error("Failed to parse NDJSON line:", line, e);
          onError?.(e as Error);
        }
      }
    }
  } catch (error) {
    console.error("Stream error:", error);
    onError?.(error as Error);
  } finally {
    reader.releaseLock();
  }
}
