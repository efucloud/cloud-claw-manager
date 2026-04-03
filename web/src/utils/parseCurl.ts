import * as curlconverter from "curlconverter";

export interface ParsedCurl {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: string | null;
}

/**
 * 安全解析 curl 命令（兼容 curlconverter v3 和 v4）
 * 不依赖 tokenize/parse，仅使用 toNode + 正则
 */
export function parseCurlCommand(curlCommand: string): ParsedCurl {
  try {
    // 清理多行 curl（移除行尾 \）
    const cleanCurl = curlCommand.replace(/\\\s*\n/g, " ");

    // 生成 Node.js 代码
    const code = curlconverter.toNode(cleanCurl);

    // 1. 提取 URL
    const urlMatch = code.match(/new URL  $ ['"`]([^'"`]+)['"`]/);
    const url = urlMatch ? urlMatch[1] : "";

    // 2. 提取 method
    let method = "GET";
    const methodMatch = code.match(/method:\s*['"`]([^'"`]+)['"`]/i);
    if (methodMatch) {
      method = methodMatch[1].toUpperCase();
    }

    // 3. 提取 headers
    const headers: Record<string, string> = {};
    const headersMatch = code.match(
      /const headers = (\{[\s\S]*?\});\s*(?:const options|console\.log)/
    );
    if (headersMatch) {
      try {
        let headerStr = headersMatch[1]
          .replace(/'/g, '"') // 'key' → "key"
          .replace(/([\w $ ]+):/g, '" $ 1":') // key: → "key":
          .replace(/\\"/g, '"') // 处理转义引号
          .replace(/,\s*}/g, "}"); // 移除尾随逗号

        const parsed = JSON.parse(headerStr);
        for (const [k, v] of Object.entries(parsed)) {
          headers[k] = String(v);
        }
      } catch (e) {
        console.warn("Failed to parse headers from generated code");
      }
    }

    // 4. 提取 body (postData)
    let body: string | null = null;
    const bodyMatch = code.match(/const postData = "(.*?)";/);
    if (bodyMatch) {
      body = bodyMatch[1]
        .replace(/\\n/g, "\n")
        .replace(/\\"/g, '"')
        .replace(/\\\\/g, "\\");
    }

    return { method, url, headers, body };
  } catch (error) {
    console.error("Parse curl error:", error);
    return {
      method: "",
      url: "",
      headers: {},
      body: null,
    };
  }
}
export function copyToClipboard(text: string) {
  if (navigator.clipboard) {
    return navigator.clipboard.writeText(text);
  } else {
    // fallback for older browsers
    const textarea = document.createElement("textarea");
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    const result = document.execCommand("copy");
    document.body.removeChild(textarea);
    return result ? Promise.resolve() : Promise.reject();
  }
}
// const handleExecute = async () => {
//   try {
//     const { method, url, headers, body } = parseCurl(curlInput);

//     const res = await request(url, {
//       method: method as any,
//       headers,
//       data: body, // umi-request 自动处理
//     });

//     setResponse(JSON.stringify(res, null, 2));
//   } catch (err) {
//     message.error("Parse or request failed: " + err.message);
//   }
// };
