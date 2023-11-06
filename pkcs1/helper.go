package pkcs1

// Cipher 列表
var CipherMap = map[string]Cipher{
    "DESCBC":     CipherDESCBC,
    "DESEDE3CBC": Cipher3DESCBC,
    "AES128CBC":  CipherAES128CBC,
    "AES192CBC":  CipherAES192CBC,
    "AES256CBC":  CipherAES256CBC,

    "DESCFB":     CipherDESCFB,
    "DESEDE3CFB": Cipher3DESCFB,
    "AES128CFB":  CipherAES128CFB,
    "AES192CFB":  CipherAES192CFB,
    "AES256CFB":  CipherAES256CFB,

    "DESOFB":     CipherDESOFB,
    "DESEDE3OFB": Cipher3DESOFB,
    "AES128OFB":  CipherAES128OFB,
    "AES192OFB":  CipherAES192OFB,
    "AES256OFB":  CipherAES256OFB,

    "DESCTR":     CipherDESCTR,
    "DESEDE3CTR": Cipher3DESCTR,
    "AES128CTR":  CipherAES128CTR,
    "AES192CTR":  CipherAES192CTR,
    "AES256CTR":  CipherAES256CTR,
}

// 获取 Cipher 类型
func GetCipherFromName(name string) Cipher {
    if data, ok := CipherMap[name]; ok {
        return data
    }

    return nil
}

// 检测 Cipher 类型
func CheckCipherFromName(name string) bool {
    if _, ok := CipherMap[name]; ok {
        return true
    }

    return false
}