package encoding

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// EncodingDetector 编码检测器
type EncodingDetector struct {
	encodings map[string]encoding.Encoding
}

// NewEncodingDetector 创建编码检测器
func NewEncodingDetector() *EncodingDetector {
	detector := &EncodingDetector{
		encodings: make(map[string]encoding.Encoding),
	}

	// 初始化支持的编码
	detector.initEncodings()

	return detector
}

// initEncodings 初始化编码映射
func (d *EncodingDetector) initEncodings() {
	// Unicode编码
	d.encodings["UTF-8"] = unicode.UTF8
	d.encodings["UTF-16"] = unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	d.encodings["UTF-16BE"] = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	d.encodings["UTF-16LE"] = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)

	// 中文编码
	d.encodings["GBK"] = simplifiedchinese.GBK
	d.encodings["GB2312"] = simplifiedchinese.HZGB2312
	d.encodings["GB18030"] = simplifiedchinese.GB18030
	d.encodings["Big5"] = traditionalchinese.Big5

	// 日文编码
	d.encodings["Shift_JIS"] = japanese.ShiftJIS
	d.encodings["EUC-JP"] = japanese.EUCJP
	d.encodings["ISO-2022-JP"] = japanese.ISO2022JP

	// 韩文编码
	d.encodings["EUC-KR"] = korean.EUCKR

	// 西欧编码
	d.encodings["ISO-8859-1"] = charmap.ISO8859_1
	d.encodings["ISO-8859-2"] = charmap.ISO8859_2
	d.encodings["ISO-8859-3"] = charmap.ISO8859_3
	d.encodings["ISO-8859-4"] = charmap.ISO8859_4
	d.encodings["ISO-8859-5"] = charmap.ISO8859_5
	d.encodings["ISO-8859-6"] = charmap.ISO8859_6
	d.encodings["ISO-8859-7"] = charmap.ISO8859_7
	d.encodings["ISO-8859-8"] = charmap.ISO8859_8
	d.encodings["ISO-8859-9"] = charmap.ISO8859_9
	d.encodings["ISO-8859-10"] = charmap.ISO8859_10
	d.encodings["ISO-8859-13"] = charmap.ISO8859_13
	d.encodings["ISO-8859-14"] = charmap.ISO8859_14
	d.encodings["ISO-8859-15"] = charmap.ISO8859_15
	d.encodings["ISO-8859-16"] = charmap.ISO8859_16

	// Windows编码
	d.encodings["Windows-1250"] = charmap.Windows1250
	d.encodings["Windows-1251"] = charmap.Windows1251
	d.encodings["Windows-1252"] = charmap.Windows1252
	d.encodings["Windows-1253"] = charmap.Windows1253
	d.encodings["Windows-1254"] = charmap.Windows1254
	d.encodings["Windows-1255"] = charmap.Windows1255
	d.encodings["Windows-1256"] = charmap.Windows1256
	d.encodings["Windows-1257"] = charmap.Windows1257
	d.encodings["Windows-1258"] = charmap.Windows1258

	// 其他编码
	d.encodings["KOI8-R"] = charmap.KOI8R
	d.encodings["KOI8-U"] = charmap.KOI8U
}

// GetSupportedEncodings 获取支持的编码列表
func (d *EncodingDetector) GetSupportedEncodings() []string {
	encodings := make([]string, 0, len(d.encodings))
	for name := range d.encodings {
		encodings = append(encodings, name)
	}
	return encodings
}

// DetectEncoding 检测编码
func (d *EncodingDetector) DetectEncoding(data []byte, calibrationText string) (string, error) {
	if calibrationText == "" {
		return "UTF-8", nil // 默认返回UTF-8
	}

	fmt.Printf("开始检测编码，校准文本: '%s'，数据长度: %d\n", calibrationText, len(data))

	// 遍历所有编码，将原始字节数据按不同编码解码
	for encodingName, enc := range d.encodings {
		decoded, err := d.decodeBytes(data, enc)
		if err != nil {
			fmt.Printf("编码 %s 解码失败: %v\n", encodingName, err)
			continue
		}

		// 检查解码后的文本是否包含校准文本
		if strings.Contains(decoded, calibrationText) {
			fmt.Printf("找到匹配编码: %s\n", encodingName)
			return encodingName, nil
		} else {
			// 显示解码后文本的前100个字符用于调试
			preview := decoded
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			fmt.Printf("编码 %s 解码成功但不包含校准文本，预览: %s\n", encodingName, preview)
		}
	}

	return "", fmt.Errorf("无法检测到包含校准文本 '%s' 的编码", calibrationText)
}

// AutoDetectEncoding 自动检测编码并转换
func (d *EncodingDetector) AutoDetectEncoding(data []byte) (string, string, error) {
	// 使用 charset 包自动检测编码
	reader := bytes.NewReader(data)

	// 尝试从内容中检测编码
	encoding, name, certain := charset.DetermineEncoding(data, "text/html")

	fmt.Printf("自动检测到编码: %s (确定性: %v)\n", name, certain)

	// 使用检测到的编码进行转换
	decoder := encoding.NewDecoder()
	transformReader := transform.NewReader(reader, decoder)

	decoded, err := io.ReadAll(transformReader)
	if err != nil {
		return "", "", fmt.Errorf("编码转换失败: %v", err)
	}

	return string(decoded), name, nil
}

// DecodeWithEncoding 使用指定编码解码
func (d *EncodingDetector) DecodeWithEncoding(data []byte, encodingName string) (string, error) {
	enc, exists := d.encodings[encodingName]
	if !exists {
		return "", fmt.Errorf("不支持的编码: %s", encodingName)
	}

	return d.decodeBytes(data, enc)
}

// decodeBytes 解码字节数据
func (d *EncodingDetector) decodeBytes(data []byte, enc encoding.Encoding) (string, error) {
	decoder := enc.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(data), decoder)

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// GetCommonEncodings 获取常用编码列表
func (d *EncodingDetector) GetCommonEncodings() []string {
	return []string{
		"UTF-8",
		"GBK",
		"GB2312",
		"GB18030",
		"Big5",
		"UTF-16",
		"UTF-16BE",
		"UTF-16LE",
		"ISO-8859-1",
		"Windows-1252",
		"Shift_JIS",
		"EUC-JP",
		"EUC-KR",
		"KOI8-R",
		"Windows-1251",
	}
}

// IsValidEncoding 检查编码是否有效
func (d *EncodingDetector) IsValidEncoding(encodingName string) bool {
	_, exists := d.encodings[encodingName]
	return exists
}
