/**
 * @name 代理的配置
 * @see 在生产环境 代理是无法生效的，所以这里没有生产环境的配置
 * -------------------------------
 * The agent cannot take effect in the production environment
 * so there is no configuration of the production environment
 * For details, please see
 * https://pro.ant.design/docs/deploy
 *
 * @doc https://umijs.org/docs/guides/proxy
 */
export default {
  dev: {
    "/api/": {
      // target: "https://openclaw.example.com/",
      target: "http://localhost:9008/",
      changeOrigin: true,
      pathRewrite: {},
      ws: true, // 全局开启 WebSocket 支持（不影响 SSE）
      // 禁用代理层的 gzip 压缩
      onProxyRes: (proxyRes, req) => {
        const url = req.url || "";
        if (url.includes("/api/stream/")) {
          // 删除任何可能的 content-encoding
          // delete proxyRes.headers["content-encoding"];
          // delete proxyRes.headers["content-length"];

          // 强制设置正确头（虽然后端已设，但保险起见）
          proxyRes.headers["cache-control"] = "no-cache";
          proxyRes.headers["connection"] = "keep-alive";
          proxyRes.headers["content-type"] =
            "application/x-ndjson; charset=utf-8";
          //  禁止代理层压缩
          // 注意：http-proxy-middleware 不直接提供开关，
          // 但可以通过移除 Accept-Encoding 来避免触发压缩
        }
      },
    },
  },
};
