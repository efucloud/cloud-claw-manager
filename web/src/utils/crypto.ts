import CryptoJS from "crypto-js";
import forge from "node-forge";

//  * @name AES-加密
//  * @param raw 待加密字段
//  * @param AESKey AES Key
//  * @return {string} 返回加密字段
/**
 * 从动态密钥派生出标准的 AES-128-CBC 所需的 Key 和 IV
 * 逻辑：SHA256(secret) -> 前16字节为Key, 后16字节为IV
 */
const deriveKeyAndIV = (secret: string) => {
  // 1. 计算 SHA256 (32字节)
  const hash = CryptoJS.SHA256(secret);

  // 2. 转为 Hex 字符串 (64个字符)
  const hex = hash.toString(CryptoJS.enc.Hex);

  // 3. 截取：前32字符(16字节)做Key，后32字符(16字节)做IV
  const keyHex = hex.substring(0, 32);
  const ivHex = hex.substring(32, 64);

  return {
    key: CryptoJS.enc.Hex.parse(keyHex),
    iv: CryptoJS.enc.Hex.parse(ivHex),
  };
};
const trimByMaxKeySize = (key: string): string => {
  return key.length > 32 ? key.substring(0, 32) : key;
};

/**
 * 对应 Go: ZerosPadding (针对 Key 的处理)
 * 如果 key 长度小于 blockSize (16)，则在末尾补 '\0' 直到长度为 16。
 * 注意：Go 的 aes.NewCipher 要求 key 长度必须是 16, 24, 或 32。
 * 如果补零后长度不满足这些值，Go 端会加密失败返回空串。
 * 这里我们严格模拟补零逻辑，假设业务场景下 key 较短，会被补到 16。
 */
const zerosPaddingKey = (key: string, blockSize: number = 16): string => {
  const len = key.length;
  if (len >= blockSize) {
    // 如果已经大于等于 16，Go 代码中 ZerosPadding 不会截断，直接返回原数组
    // 但后续 aes.NewCipher 会校验长度。这里保持原样，由 CryptoJS 处理或模拟 Go 的报错行为
    return key;
  }
  return key + "\0".repeat(blockSize - len);
};

/**
 * 对应 Go: GenIVFromKey
 * 1. SHA256(key)
 * 2. 转为 Hex 字符串
 * 3. 截取前 16 个字符 (trimByBlockSize)
 * 4. 转为字节数组 (在 JS 中即 parse 为 WordArray)
 */
const genIVFromKey = (key: string): CryptoJS.lib.WordArray => {
  // 注意：Go 中 GenIVFromKey 接收的是 trimByMaxKeySize 之后的 key
  const hashHex = CryptoJS.SHA256(key).toString(CryptoJS.enc.Hex);
  // 截取前 16 个字符
  const ivString = hashHex.substring(0, 16);
  // 将这 16 个 ASCII 字符解析为字节
  return CryptoJS.enc.Utf8.parse(ivString);
};

/**
 * 对应 Go: AesSimpleEncrypt
 * @param data 原始数据字符串
 * @param key 密钥字符串
 * @returns Base64 密文 (如果失败返回空字符串，与 Go 一致)
 */
export const aesEncrypt = (data: string, key: string): string => {
  try {
    // 1. 处理 Key: trim(32) -> ZerosPadding(16)
    let processedKey = trimByMaxKeySize(key);
    processedKey = zerosPaddingKey(processedKey, 16);

    // 2. 生成 IV: 使用 trim 之后、padding 之前的 key (参考 Go 代码调用顺序)
    // Go: key = trim...; keyBytes = Zeros...; GenIVFromKey(key)
    // 注意：Go 代码中 GenIVFromKey(key) 用的是 trim 后的 key，不是 padding 后的 keyBytes
    const ivSourceKey = trimByMaxKeySize(key);
    const iv = genIVFromKey(ivSourceKey);

    // 3. 准备 Key WordArray
    const keyWordArray = CryptoJS.enc.Utf8.parse(processedKey);

    // 4. 加密配置
    const cfg = {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7, // 对应 Go 的 PKCS7
    };

    const encrypted = CryptoJS.AES.encrypt(data, keyWordArray, cfg);
    return encrypted.toString(); // 默认输出 Base64
  } catch (e) {
    console.error("AesSimpleEncrypt error:", e);
    return "";
  }
};

/**
 * 对应 Go: AesSimpleDecrypt
 * @param base64Cipher Base64 密文
 * @param key 密钥字符串
 * @returns 原始数据字符串 (如果失败返回空字符串，与 Go 一致)
 */
export const aesDecrypt = (base64Cipher: string, key: string): string => {
  try {
    // 1. 处理 Key: trim(32) -> ZerosPadding(16)
    let processedKey = trimByMaxKeySize(key);
    processedKey = zerosPaddingKey(processedKey, 16);

    // 2. 生成 IV
    const ivSourceKey = trimByMaxKeySize(key);
    const iv = genIVFromKey(ivSourceKey);

    // 3. 准备 Key WordArray
    const keyWordArray = CryptoJS.enc.Utf8.parse(processedKey);

    // 4. 解密配置
    const cfg = {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7,
    };

    const decrypted = CryptoJS.AES.decrypt(base64Cipher, keyWordArray, cfg);

    // 5. 转为 UTF8 字符串
    const result = decrypted.toString(CryptoJS.enc.Utf8);

    // Go 代码中如果解密失败或结果为空，通常返回 "" (由上层判断)
    // 如果 decrypt 对象为空或 stringify 失败，这里可能会得到空串
    return result;
  } catch (e) {
    console.error("AesSimpleDecrypt error:", e);
    return "";
  }
};
export const decodeBase64 = (base64: string): string => {
  const binString = atob(base64);
  return new TextDecoder().decode(
    Uint8Array.from(binString, (c) => c.charCodeAt(0))
  );
};

/**
 * 解析 PEM 格式的 X.509 证书，返回有效期信息
 * @param {string} pemCert - PEM 格式的证书字符串（包含 -----BEGIN CERTIFICATE-----）
 * @returns {{
 *   valid: boolean,
 *   notBefore: Date | null,
 *   notAfter: Date | null,
 *   isExpired: boolean,
 *   daysUntilExpiry: number | null,
 *   issuer: string | null,
 *   subject: string | null,
 *   error?: string
 * }}
 */
export const parseCertificateExpiry = (pemCert: string) => {
  const result = {
    valid: false,
    notBefore: null as Date | null,
    notAfter: null as Date | null,
    isExpired: false,
    daysUntilExpiry: null as number | null,
    issuer: null as string | null,
    subject: null as string | null,
    error: null as string | null,
  };

  try {
    if (typeof forge === "undefined") {
      result.error = "forge is not available";
      return result;
    }

    // 清理输入：移除前后空白，确保格式统一
    const cleanPem = pemCert.trim();

    // 提取第一个有效的证书块（以 BEGIN/END 包裹）
    const certMatch = cleanPem.match(
      /-----BEGIN CERTIFICATE-----\s*[\s\S]*?\s*-----END CERTIFICATE-----/
    );

    if (!certMatch) {
      result.error = "No valid certificate block found in PEM";
      return result;
    }

    const firstCertPem = certMatch[0];

    // 解析第一个证书
    const cert = forge.pki.certificateFromPem(firstCertPem);

    const notBefore = new Date(cert.validity.notBefore);
    const notAfter = new Date(cert.validity.notAfter);
    const now = new Date();

    const isExpired = notAfter < now;
    const timeDiff = notAfter.getTime() - now.getTime();
    const daysUntilExpiry =
      timeDiff > 0 ? Math.ceil(timeDiff / (1000 * 60 * 60 * 24)) : 0;

    result.valid = true;
    result.notBefore = notBefore;
    result.notAfter = notAfter;
    result.isExpired = isExpired;
    result.daysUntilExpiry = daysUntilExpiry;
    result.issuer = forge.pki.DN.stringify(cert.issuer);
    result.subject = forge.pki.DN.stringify(cert.subject);
  } catch (err: any) {
    result.error = err?.message || "Failed to parse certificate";
    console.error("[parseCertificateExpiry] Error:", err);
  }

  return result;
};
/**
 * 计算目标日期距离当前时间还有多少天（向上取整）
 * 常用于计算证书/授权等的剩余有效期
 *
 * @param {Date} expiryDate - 过期时间（UTC 或本地 Date 对象）
 * @param {Date} [now=new Date()] - 当前时间（可选，默认为当前时间）
 * @returns {number}
 *   - 正数：还剩多少天（例如 5 表示 5 天后过期）
 *   - 0：今天过期
 *   - 负数：已过期多少天（例如 -3 表示 3 天前已过期）
 */
export const getDaysUntilExpiry = (expiryDate, now = new Date()) => {
  if (!(expiryDate instanceof Date) || isNaN(expiryDate.getTime())) {
    console.log("Invalid expiryDate: must be a valid Date object");
    return 0;
  }
  if (!(now instanceof Date) || isNaN(now.getTime())) {
    console.log("Invalid now: must be a valid Date object");
    return 0;
  }

  const timeDiff = expiryDate.getTime() - now.getTime();
  const days = Math.ceil(timeDiff / (1000 * 60 * 60 * 24));
  return days;
};
