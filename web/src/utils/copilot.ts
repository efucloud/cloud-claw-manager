/**
 * 清理 Kubernetes 资源对象，仅移除以下两个字段：
 * - metadata.managedFields
 * - metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']
 *
 * @param resource 原始 Kubernetes 资源对象（如 Pod、Deployment 等）
 * @returns 清理后的深拷贝对象（不修改原对象）
 */
export function cleanK8sResourceForAI(resource: any): any {
  const cleaned = JSON.parse(JSON.stringify(resource));

  if (cleaned?.metadata && typeof cleaned.metadata === "object") {
    // 删除 managedFields
    delete cleaned.metadata.managedFields;

    // 删除 kubectl 最后应用配置
    if (cleaned.metadata.annotations) {
      delete cleaned.metadata.annotations[
        "kubectl.kubernetes.io/last-applied-configuration"
      ];
    }
  }

  return cleaned;
}
export function nodesEdgesToMermaid(nodes, edges, direction = "TD") {
  const nodeLines = [];
  const edgeLines = [];

  // 映射节点类型到 Mermaid 形状
  const getShape = (type, label) => {
    if (type === "input") return `[ ${label}]`; // 矩形（默认）
    if (type === "output") return `[[ ${label}]]`; // 圆角矩形
    if (type === "decision") return `{ ${label}}`; // 菱形（判断）
    return `[ ${label}]`; // 默认矩形
  };

  // 生成节点行
  for (const node of nodes) {
    const id = node.id.replace(/[^a-zA-Z0-9_-]/g, "_"); // 清理非法字符
    const label = node.data?.label || node.id;
    const shape = getShape(node.type, label);
    nodeLines.push(`     ${id} ${shape}`);
  }

  // 生成边行
  for (const edge of edges) {
    const source = edge.source.replace(/[^a-zA-Z0-9_-]/g, "_");
    const target = edge.target.replace(/[^a-zA-Z0-9_-]/g, "_");
    const label = edge.label ? `| ${edge.label}|` : "";
    edgeLines.push(`     ${source} --> ${label}  ${target}`);
  }

  return `graph  ${direction}\n ${nodeLines.join('\n')}\n ${edgeLines.join('\n')}`;
}
