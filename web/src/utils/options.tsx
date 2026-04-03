import { FormattedMessage } from "@umijs/max";

export const clusterVolumeProviderOptions = [
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.secret" />,
    value: "secret"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.configMap" />,
    value: "configMap"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.emptyDir" />,
    value: "emptyDir"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.persistentVolumeClaim" />,
    value: "persistentVolumeClaim"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.hostPath" />,
    value: "hostPath"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.nfs" />,
    value: "nfs"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.ephemeral" />,
    value: "ephemeral"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.csi" />,
    value: "csi"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.downwardAPI" />,
    value: "downwardAPI"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.fc" />,
    value: "fc"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.image" />,
    value: "image"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.iscsi" />,
    value: "iscsi"
  },
  {
    label: <FormattedMessage id="cluster.resource.volume.provider.projected" />,
    value: "projected"
  },
];

export const capabilitiesOptions = [
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.ALL" defaultMessage="所有(ALL)" />,
    value: "ALL"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.AUDIT_CONTROL" defaultMessage="审计控制(AUDIT_CONTROL)" />,
    value: "AUDIT_CONTROL"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.AUDIT_READ" defaultMessage="审计读取(AUDIT_READ)" />,
    value: "AUDIT_READ"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.AUDIT_WRITE" defaultMessage="审计写入(AUDIT_WRITE)" />,
    value: "AUDIT_WRITE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.BLOCK_SUSPEND" defaultMessage="阻止系统挂起(BLOCK_SUSPEND)" />,
    value: "BLOCK_SUSPEND"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.BPF" defaultMessage="BPF 程序操作(BPF)" />,
    value: "BPF"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.CHECKPOINT_RESTORE" defaultMessage="检查点/恢复(CHECKPOINT_RESTORE)" />,
    value: "CHECKPOINT_RESTORE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.CHOWN" defaultMessage="更改文件所有者(CHOWN)" />,
    value: "CHOWN"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.DAC_OVERRIDE" defaultMessage="绕过文件权限检查(DAC_OVERRIDE)" />,
    value: "DAC_OVERRIDE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.DAC_READ_SEARCH" defaultMessage="绕过文件读/搜索限制(DAC_READ_SEARCH)" />,
    value: "DAC_READ_SEARCH"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.FOWNER" defaultMessage="忽略文件属主限制(FOWNER)" />,
    value: "FOWNER"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.FSETID" defaultMessage="保留 setuid/setgid 位(FSETID)" />,
    value: "FSETID"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.IPC_LOCK" defaultMessage="锁定内存(防交换)(IPC_LOCK)" />,
    value: "IPC_LOCK"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.IPC_OWNER" defaultMessage="绕过 IPC 权限检查(IPC_OWNER)" />,
    value: "IPC_OWNER"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.KILL" defaultMessage="发送任意信号(KILL)" />,
    value: "KILL"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.LEASE" defaultMessage="建立文件租约(LEASE)" />,
    value: "LEASE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.LINUX_IMMUTABLE" defaultMessage="设置文件不可变属性(LINUX_IMMUTABLE)" />,
    value: "LINUX_IMMUTABLE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.MAC_ADMIN" defaultMessage="MAC 策略管理(MAC_ADMIN)" />,
    value: "MAC_ADMIN"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.MAC_OVERRIDE" defaultMessage="绕过 MAC 策略(MAC_OVERRIDE)" />,
    value: "MAC_OVERRIDE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.MKNOD" defaultMessage="创建设备节点(MKNOD)" />,
    value: "MKNOD"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.NET_ADMIN" defaultMessage="网络管理(NET_ADMIN)" />,
    value: "NET_ADMIN"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.NET_BIND_SERVICE" defaultMessage="绑定特权端口(NET_BIND_SERVICE)" />,
    value: "NET_BIND_SERVICE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.NET_BROADCAST" defaultMessage="网络广播(NET_BROADCAST)" />,
    value: "NET_BROADCAST"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.NET_RAW" defaultMessage="原始套接字(NET_RAW)" />,
    value: "NET_RAW"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.PERFMON" defaultMessage="性能监控(PERFMON)" />,
    value: "PERFMON"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SETGID" defaultMessage="修改 GID(SETGID)" />,
    value: "SETGID"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SETFCAP" defaultMessage="设置文件能力(SETFCAP)" />,
    value: "SETFCAP"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SETPCAP" defaultMessage="修改进程能力(SETPCAP)" />,
    value: "SETPCAP"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SETUID" defaultMessage="修改 UID(SETUID)" />,
    value: "SETUID"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_ADMIN" defaultMessage="系统管理(极高危)(SYS_ADMIN)" />,
    value: "SYS_ADMIN"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_BOOT" defaultMessage="系统启动/关机(SYS_BOOT)" />,
    value: "SYS_BOOT"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_CHROOT" defaultMessage="使用 chroot(SYS_CHROOT)" />,
    value: "SYS_CHROOT"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_MODULE" defaultMessage="加载/卸载内核模块(极度危险)(SYS_MODULE)" />,
    value: "SYS_MODULE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_NICE" defaultMessage="调整进程优先级(SYS_NICE)" />,
    value: "SYS_NICE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_PACCT" defaultMessage="启用进程审计(SYS_PACCT)" />,
    value: "SYS_PACCT"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_PTRACE" defaultMessage="进程调试/跟踪(SYS_PTRACE)" />,
    value: "SYS_PTRACE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_RAWIO" defaultMessage="原始 I/O(危险)(SYS_RAWIO)" />,
    value: "SYS_RAWIO"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_RESOURCE" defaultMessage="修改资源限制(SYS_RESOURCE)" />,
    value: "SYS_RESOURCE"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_TIME" defaultMessage="修改系统时间(SYS_TIME)" />,
    value: "SYS_TIME"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYS_TTY_CONFIG" defaultMessage="TTY 配置(SYS_TTY_CONFIG)" />,
    value: "SYS_TTY_CONFIG"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.SYSLOG" defaultMessage="读取内核日志(SYSLOG)" />,
    value: "SYSLOG"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.WAKE_ALARM" defaultMessage="设置唤醒定时器(WAKE_ALARM)" />,
    value: "WAKE_ALARM"
  },
  {
    label: <FormattedMessage id="cluster.resource.container.form.security.capability.WINESYNC" defaultMessage="Wine 同步原语(WINESYNC)" />,
    value: "WINESYNC"
  }
];