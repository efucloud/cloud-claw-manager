// utils/imageDownload.ts

import { getToken } from "./global";

export async function namespaceImageDownload(params: {
  cluster: string;
  namespace: string;
  image?: string;
}) {
  const { cluster, namespace, image } = params;
  const orgToken = getToken();

  const url = new URL(
    `/api/v1/cluster/${cluster}/namespace/${namespace}/image/download`,
    window.location.origin
  );

  if (image) {
    url.searchParams.set("image", image);
  }

  try {
    const response = await fetch(url.toString(), {
      method: "GET",
      headers: {
        Authorization: `Bearer ${orgToken.access_token}`, // 👈 关键：注入 Token
        // 注意：不要设置 Content-Type，GET 请求无 body
      },
    });

    if (!response.ok) {
      // 尝试读取错误信息（可能是 JSON 或文本）
      let errorMessage = `下载失败: ${response.status} ${response.statusText}`;
      try {
        const errorText = await response.text();
        if (errorText) {
          // 如果是 JSON 错误
          try {
            const errorJson = JSON.parse(errorText);
            errorMessage = errorJson.message || errorJson.error || errorText;
          } catch {
            errorMessage = errorText;
          }
        }
      } catch (e) {
        // 忽略读取错误
      }
      throw new Error(errorMessage);
    }

    // 获取文件名
    const contentDisposition = response.headers.get("Content-Disposition");
    let filename = "image.tar";
    if (contentDisposition) {
      // 支持 RFC 5987 UTF-8 文件名（如 filename*=UTF-8''xxx.tar）
      const utf8Filename = contentDisposition.match(/filename\*=UTF-8''(.+)/i);
      if (utf8Filename) {
        filename = decodeURIComponent(utf8Filename[1]);
      } else {
        // 兼容 ASCII 文件名（如 filename="xxx.tar"）
        const asciiFilename = contentDisposition.match(
          /filename="?([^;"]+)"?/i
        );
        if (asciiFilename) {
          filename = asciiFilename[1];
        }
      }
    }

    // 获取 Blob 并触发下载
    const blob = await response.blob();

    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = downloadUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  } catch (error) {
    console.error("Image download error:", error);
    throw error; // 让调用方可以捕获并提示用户
  }
}
