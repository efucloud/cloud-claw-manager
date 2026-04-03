// utils/markdownToPdf.ts
import MarkdownIt from 'markdown-it';
import hljs from 'highlight.js';
import 'highlight.js/styles/github.css'; // 代码高亮样式
import jsPDF from 'jspdf';
import html2canvas from 'html2canvas';

/**
 * 将 Markdown 字符串导出为 PDF
 * @param markdown - 原始 Markdown 文本
 * @param filename - 输出文件名（默认: 'document.pdf'）
 * @param options - 渲染选项
 */
export const exportMarkdownToPdf = async (
  markdown: string,
  filename = 'document.pdf',
  options?: {
    /** 页面宽度（px），默认 800 */
    width?: number;
    /** 字体大小（px），默认 16 */
    fontSize?: number;
    /** 是否启用代码高亮 */
    highlightCode?: boolean;
  }
): Promise<void> => {
  const {
    width = 800,
    fontSize = 16,
    highlightCode = true,
  } = options || {};

  // Step 1: Markdown → HTML
  const md = new MarkdownIt({
    html: false,        // 禁用原始 HTML（安全）
    linkify: true,      // 自动识别 URL
    typographer: true,  // 启用引号等排版优化
    highlight: highlightCode
      ? (str, lang) => {
          if (lang && hljs.getLanguage(lang)) {
            try {
              return hljs.highlight(str, { language: lang }).value;
            } catch (__) {}
          }
          return ''; // 使用 highlight.js 默认样式
        }
      : undefined,
  });

  const html = md.render(markdown);

  // Step 2: 创建临时 DOM 容器
  const container = document.createElement('div');
  container.innerHTML = html;
  container.style.width = `${width}px`;
  container.style.padding = '24px';
  container.style.fontFamily = '"Segoe UI", "Microsoft YaHei", sans-serif';
  container.style.fontSize = `${fontSize}px`;
  container.style.lineHeight = '1.6';
  container.style.color = '#333';
  container.style.backgroundColor = 'white';
  container.style.position = 'fixed';
  container.style.left = '-9999px'; // 隐藏但保持渲染
  container.style.top = '0';
  container.style.zIndex = '-1';

  // 添加基础样式（确保表格/代码块显示正常）
  const style = document.createElement('style');
  style.textContent = `
    body { margin: 0; }
    pre { 
      background: #f6f8fa; 
      padding: 16px; 
      border-radius: 6px; 
      overflow-x: auto;
    }
    code { 
      font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
      font-size: ${Math.max(12, fontSize - 2)}px;
    }
    table { 
      border-collapse: collapse; 
      width: 100%; 
      margin: 16px 0;
    }
    th, td { 
      border: 1px solid #d0d7de; 
      padding: 8px 12px; 
      text-align: left;
    }
    blockquote {
      margin: 0;
      padding-left: 16px;
      border-left: 4px solid #d0d7de;
      color: #57606a;
    }
    img {
      max-width: 100%;
      height: auto;
    }
  `;
  container.appendChild(style);

  document.body.appendChild(container);

  try {
    // Step 3: HTML → Canvas
    const canvas = await html2canvas(container, {
      scale: 2,
      useCORS: true,
      allowTaint: true,
      backgroundColor: '#ffffff',
    });

    // Step 4: Canvas → PDF
    const imgData = canvas.toDataURL('image/png');
    const pdf = new jsPDF({
      orientation: canvas.width > canvas.height ? 'landscape' : 'portrait',
      unit: 'px',
      format: [canvas.width, canvas.height],
    });

    pdf.addImage(imgData, 'PNG', 0, 0, canvas.width, canvas.height);
    pdf.save(filename);
  } finally {
    // 清理临时 DOM
    document.body.removeChild(container);
  }
};