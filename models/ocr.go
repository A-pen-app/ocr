package models

import (
	"time"
)

type OCRRawInfo struct {
	IdentifyURL        *string `json:"identify_url,omitempty"`
	Name               *string `json:"name"`
	Birthday           *string `json:"birthday"`
	Position           *string `json:"position,omitempty"`
	Department         *string `json:"department,omitempty"`
	Facility           *string `json:"facility,omitempty"`
	ValidDate          *string `json:"valid_date,omitempty"`
	SpecialtyValidDate *string `json:"specialty_valid_date,omitempty"`
}

type OCRInfo struct {
	Name       *string `json:"name"`
	Position   *string `json:"position"`
	Department *string `json:"department"`
	Facility   *string `json:"facility"`
}

type OCREventMessage struct {
	UserID    string    `json:"user_id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
}

type OCRMessageType string

const (
	OCRMessageTypeIdentifyOCR OCRMessageType = "identify_ocr"
)

type OCRTopic string

const (
	OCRTopicDev  OCRTopic = "wanderer-dev"
	OCRTopicProd OCRTopic = "wanderer-prod"
)

type PlatformType string

const (
	PlatformTypeApen  PlatformType = "apen"
	PlatformTypeNurse PlatformType = "nurse"
	PlatformTypePhar  PlatformType = "phar"
)

func GetInfoPrompt(professionType PlatformType) string {
	switch professionType {
	case PlatformTypeApen:
		return apenInfoPrompt
	case PlatformTypeNurse:
		return nurseInfoPrompt
	case PlatformTypePhar:
		return pharInfoPrompt
	default:
		return apenInfoPrompt // default to apen
	}
}

const SystemContent = "You are a helpful assistant that analyzes images and outputs information with JSON format."

const NamePrompt = `
這是一張參加證、識別證、執照、證書、或名片，請判斷其中的中文姓名，並以以下 JSON 格式輸出：
{
  "name": "中文姓名"
}
如果找不到中文姓名，請將 "name" 的值設為空字串。
	`

const apenInfoPrompt = `
請分析這張圖片（可能是醫師的識別證、執照、證書或名片），並提取以下資訊：

**需要辨識的欄位：**

1. **name（姓名）**: 中文姓名
2. **birthday（生日）**: 格式為 YYYY-MM-DD
3. **department（科別）**: 中文科別名稱
4. **facility（執業場所）**: 任職場所或執業場所
   - 注意：學生不需要填寫此欄位
5. **position（職級）**: 醫師職級，僅限以下三種：
   - "PGY" - 不分科醫師
   - "Resident" - 住院醫師
   - "VS" - 主治醫師（專科證書必定為 VS）
   - 如果只標註「醫師」而無具體職級，或為學生則不填
6. **valid_date（醫師證書生效日期）**: 格式為 YYYY-MM-DD
   - 僅當圖片為「醫師證書」時才需辨識
   - 執業執照不需要填寫此欄位
7. **specialty_valid_date（專科證書生效日期）**: 格式為 YYYY-MM-DD
   - 僅當圖片為「專科證書」時才需辨識
   - 注意：這是「生效日期」或「頒發日期」，不是「有效日期」

**輸出格式：**
請以以下 JSON 格式輸出（如果找不到對應資料或無法辨識，請將該欄位的值設為 null）：

{
  "name": "中文姓名",
  "birthday": "YYYY-MM-DD",
  "position": "PGY、Resident 或 VS",
  "department": "科別",
  "facility": "執業場所/任職場所",
  "valid_date": "YYYY-MM-DD",
  "specialty_valid_date": "YYYY-MM-DD"
}
	`

const nurseInfoPrompt = `
請分析這張圖片（可能是護理師的識別證、執照、證書或名片），並提取以下資訊：

**需要辨識的欄位：**

1. **name（姓名）**: 中文姓名
2. **birthday（生日）**: 格式為 YYYY-MM-DD
3. **department（科別）**: 中文科別名稱
4. **facility（執業場所）**: 任職場所或執業場所
   - 注意：學生不需要填寫此欄位
5. **valid_date（護理師證書生效日期）**: 格式為 YYYY-MM-DD
   - 僅當圖片為「護理師證書」時才需辨識
   - 執業執照不需要填寫此欄位

**輸出格式：**
請以以下 JSON 格式輸出（如果找不到對應資料或無法辨識，請將該欄位的值設為 null）：

{
  "name": "中文姓名",
  "birthday": "YYYY-MM-DD",
  "department": "科別",
  "facility": "執業場所/任職場所",
  "valid_date": "YYYY-MM-DD"
}
	`

const pharInfoPrompt = `
請分析這張圖片（可能是藥師的識別證、執照、證書或名片），並提取以下資訊：

**需要辨識的欄位：**

1. **name（姓名）**: 中文姓名
2. **birthday（生日）**: 格式為 YYYY-MM-DD
3. **facility（執業場所）**: 任職場所或執業場所
   - 注意：學生不需要填寫此欄位
4. **valid_date（藥師證書生效日期）**: 格式為 YYYY-MM-DD
   - 僅當圖片為「藥師證書」時才需辨識
   - 執業執照不需要填寫此欄位

**輸出格式：**
請以以下 JSON 格式輸出（如果找不到對應資料或無法辨識，請將該欄位的值設為 null）：

{
  "name": "中文姓名",
  "birthday": "YYYY-MM-DD",
  "facility": "執業場所/任職場所",
  "valid_date": "YYYY-MM-DD"
}
	`
